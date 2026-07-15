"use client";

import { usePathname, useRouter } from "next/navigation";
import { parseAsBoolean, useQueryState } from "nuqs";
import { useState } from "react";

import { Permission, ThreadReference } from "src/api/openapi-schema";

import { handle } from "@/api/client";
import { threadSharePin } from "@/api/openapi-client/channels";
import { useSession } from "@/auth";
import { useConfirmation } from "@/components/site/useConfirmation";
import { useFeedMutations } from "@/lib/feed/mutation";
import { useReportContext } from "@/lib/report/useReportContext";
import { canDeletePost, canEditPost } from "@/lib/thread/permissions";
import { withUndo } from "@/lib/thread/undo";
import { useShare } from "@/utils/client";
import { hasPermission } from "@/utils/permissions";
import { useCopyToClipboard } from "@/utils/useCopyToClipboard";

import { getPermalinkForThread } from "../utils";

export type Props = {
  thread: ThreadReference;
  editingEnabled?: boolean;
  movingEnabled?: boolean;
  onPinChange?: (pinned: boolean) => Promise<void>;
  channelID?: string;
};

export function useThreadMenu({
  thread,
  editingEnabled,
  movingEnabled,
  onPinChange,
  channelID,
}: Props) {
  const router = useRouter();
  const account = useSession();
  const { resolveReport } = useReportContext();
  const [_, setEditing] = useQueryState("edit", parseAsBoolean);
  const [, copyToClipboard] = useCopyToClipboard();
  const pathname = usePathname();
  const isOnThreadPage = pathname?.includes(`/t/${thread.slug}`);

  const { deleteThread, updateThread, revalidate } = useFeedMutations(undefined, undefined, undefined, router);

  const {
    isConfirming: isConfirmingDelete,
    handleConfirmAction: handleConfirmDelete,
    handleCancelAction: handleCancelDelete,
  } = useConfirmation(handleDelete);

  const isSharingEnabled = useShare();
  const isEditingEnabled = canEditPost(thread, account) && editingEnabled;
  const isMovingEnabled = canEditPost(thread, account) && movingEnabled;
  const isDeletingEnabled =
    canDeletePost(thread, account) && thread.deletedAt === undefined;
  const canPinThread = hasPermission(account, Permission.MANAGE_POSTS);
  const isThreadPinned = (thread.pinned ?? 0) > 0;

  const canShareThread = hasPermission(account, Permission.ADMINISTRATOR);
  const canPinInChannel =
    canShareThread && !!channelID && !!thread.shared_from;
  const [isPinnedInChannel, setIsPinnedInChannel] = useState(
    thread.shared_from !== undefined && (thread.pinned ?? 0) > 0,
  );

  const permalink = getPermalinkForThread(thread.slug);

  async function handleCopyLink() {
    copyToClipboard(permalink);
  }

  async function handleShare() {
    await navigator.share({
      title: `A post by ${thread.author.name}`,
      url: permalink,
      text: thread.description,
    });
  }

  function handleEdit() {
    setEditing(true);
  }

  async function handleDelete() {
    await handle(async () => {
      await withUndo({
        message: "Thread deleted",
        duration: 5000,
        toastId: `thread-${thread.id}`,
        action: async () => {
          await deleteThread(thread.id);
          await resolveReport();

          if (isOnThreadPage) {
            router.push("/");
          }
        },
        onUndo: () => {},
      });
    });
  }

  async function handlePinThread() {
    await handle(
      async () => {
        if (onPinChange) {
          await onPinChange(true);
        } else {
          await updateThread(thread.id, { pinned: 1 });
        }
      },
      {
        cleanup: async () => await revalidate(),
      },
    );
  }

  async function handleUnpinThread() {
    await handle(
      async () => {
        if (onPinChange) {
          await onPinChange(false);
        } else {
          await updateThread(thread.id, { pinned: 0 });
        }
      },
      {
        cleanup: async () => await revalidate(),
      },
    );
  }

  async function handlePinInChannel() {
    if (!channelID) return;

    await handle(
      async () => {
        await threadSharePin(channelID, thread.slug, { pinned: true });
        setIsPinnedInChannel(true);
      },
      {
        cleanup: async () => await revalidate(),
      },
    );
  }

  async function handleUnpinInChannel() {
    if (!channelID) return;

    await handle(
      async () => {
        await threadSharePin(channelID, thread.slug, { pinned: false });
        setIsPinnedInChannel(false);
      },
      {
        cleanup: async () => await revalidate(),
      },
    );
  }

  return {
    isSharingEnabled,
    isEditingEnabled,
    isMovingEnabled,
    isDeletingEnabled,
    isConfirmingDelete,
    canPinThread,
    isThreadPinned,
    canShareThread,
    canPinInChannel,
    isPinnedInChannel,
    handlers: {
      handleCopyLink,
      handleShare,
      handleEdit,
      handleConfirmDelete,
      handleCancelDelete,
      handlePinThread,
      handleUnpinThread,
      handlePinInChannel,
      handleUnpinInChannel,
    },
  };
}
