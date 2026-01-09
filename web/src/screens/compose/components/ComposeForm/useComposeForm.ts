import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useSWRConfig } from "swr";
import { z } from "zod";

import {
  channelThreadCreate,
  getChannelThreadListKey,
} from "src/api/openapi-client/channels";
import { threadCreate, threadUpdate } from "src/api/openapi-client/threads";
import { Thread, ThreadInitialProps, Visibility, Permission } from "src/api/openapi-schema";

import { handle } from "@/api/client";
import { NO_CATEGORY_VALUE } from "@/components/category/CategorySelect/useCategorySelect";
import { hasPermission } from "@/utils/permissions";
import { useSession } from "@/auth";

export type Props = {
  editing?: string;
  initialDraft?: Thread;
  channelID?: string;
  categoryID?: string;
  onSuccess?: () => void;
};

export const FormShapeSchema = z.object({
  title: z.string().min(1, "Your post must have a title to be published"),
  body: z.string().min(1),
  category: z.string().optional(),
  tags: z.string().array().optional(),
  url: z.string().optional(),
});
export type FormShape = z.infer<typeof FormShapeSchema>;

export function useComposeForm({
  initialDraft,
  editing,
  channelID,
  categoryID,
  onSuccess,
}: Props) {
  const router = useRouter();
  const { mutate } = useSWRConfig();
  const session = useSession();

  const [isPublishing, setIsPublishing] = useState(false);
  const [isSavingDraft, setIsSavingDraft] = useState(false);

  const form = useForm<FormShape>({
    resolver: zodResolver(FormShapeSchema),
    reValidateMode: "onChange",
    defaultValues: initialDraft
      ? {
          title: initialDraft.title,
          body: initialDraft.body,
          tags: initialDraft.tags.map((t) => t.name),
          url: initialDraft.link?.url,
          category: initialDraft.category?.id,
        }
      : {
          title: "",
          body: "",
          category: categoryID,
        },
  });

  const saveDraft = async (data: FormShape) => {
    const payload: ThreadInitialProps = {
      ...data,

      // When saving a new draft, these are optional but must be explicitly set.
      title: data.title ?? "",
      body: data.body ?? "",
      url: data.url ?? "",
      tags: data.tags ?? [],
      category: data.category === NO_CATEGORY_VALUE ? undefined : data.category,

      visibility: Visibility.draft,
    };

    if (editing) {
      await threadUpdate(editing, payload);
    } else {
      if (channelID) {
        const { id } = await channelThreadCreate(channelID, payload);
        router.push(`/new?id=${id}&channel=${channelID}`);
      } else {
        const { id } = await threadCreate(payload);
        router.push(`/new?id=${id}`);
      }
    }
  };

  const publish = async ({ title, body, category, tags, url }: FormShape) => {
    const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);
    const targetVisibility = isAdmin ? Visibility.published : Visibility.review;

    const payload = {
      title,
      body,
      category: category === NO_CATEGORY_VALUE ? undefined : category,
      visibility: targetVisibility,
      tags,
      url,
    };

    if (editing) {
      const { slug } = await threadUpdate(editing, payload);
      if (targetVisibility === Visibility.review) {
        router.back();
      } else {
        router.push(`/t/${slug}`);
      }
    } else {
      if (channelID) {
        await channelThreadCreate(channelID, payload);
        const threadListKey = getChannelThreadListKey(channelID, {});
        await mutate(threadListKey);
        if (onSuccess) {
          onSuccess();
        } else {
          if (targetVisibility === Visibility.review) {
            router.push(`/channels/${channelID}`);
          } else {
            router.push(`/channels/${channelID}`);
          }
        }
      } else {
        const { slug } = await threadCreate(payload);
        if (targetVisibility === Visibility.review) {
          router.back();
        } else {
          router.push(`/t/${slug}`);
        }
      }
    }
  };

  const handleSaveDraft = form.handleSubmit((data) =>
    handle(
      async () => {
        setIsSavingDraft(true);
        await saveDraft(data);
      },
      {
        promiseToast: {
          loading: "Saving draft...",
          success: "Draft saved!",
        },
        cleanup: async () => {
          setIsSavingDraft(false);
        },
      },
    ),
  );

  const handlePublish = form.handleSubmit((data) =>
    handle(
      async () => {
        setIsPublishing(true);
        await publish(data);
      },
      {
        promiseToast: {
          loading: session && hasPermission(session, Permission.ADMINISTRATOR)
            ? "Publishing post..."
            : "Submitting for review...",
          success: session && hasPermission(session, Permission.ADMINISTRATOR)
            ? "Post published!"
            : "Submitted for review! Your post will be visible once approved by a moderator.",
        },
        cleanup: async () => {
          setIsPublishing(false);
        },
      },
    ),
  );

  const handleAssetUpload = async () => {
    await handle(
      async () => {
        setIsSavingDraft(true);
        const state = form.getValues();
        await saveDraft(state);
      },
      {
        promiseToast: {
          loading: "Saving draft...",
          success: "Draft saved!",
        },
        cleanup: async () => {
          setIsSavingDraft(false);
        },
      },
    );
  };

  const handleAssetDelete = async () => {
    await handle(
      async () => {
        setIsSavingDraft(true);
        const state = form.getValues();
        await saveDraft(state);
      },
      {
        promiseToast: {
          loading: "Saving draft...",
          success: "Draft saved!",
        },
        cleanup: async () => {
          setIsSavingDraft(false);
        },
      },
    );
  };

  function handleBack() {
    router.back();
  }

  return {
    form,
    state: {
      isPublishing,
      isSavingDraft,
    },
    handlers: {
      handleSaveDraft,
      handlePublish,
      handleAssetDelete,
      handleAssetUpload,
      handleBack,
    },
  };
}
