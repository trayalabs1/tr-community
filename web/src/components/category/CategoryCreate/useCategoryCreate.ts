import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { mutate } from "swr";
import { z } from "zod";

import {
  categoryCreate,
  getCategoryListKey,
} from "src/api/openapi-client/categories";
import {
  channelCategoryCreate,
  getChannelCategoryListKey,
} from "src/api/openapi-client/channels";
import { Asset } from "src/api/openapi-schema";
import { UseDisclosureProps } from "src/utils/useDisclosure";

import { handle } from "@/api/client";

export const FormSchema = z.object({
  name: z.string().min(1, "Please enter a name for the category."),
  slug: z.string().min(1, "Please enter a URL slug for the category."),
  description: z.string().min(1, "Please enter a short description."),
  colour: z.string().default("#8577ce"),
  parent: z.string().optional(),
  cover_image_asset_id: z.string().optional(),
});
export type Form = z.infer<typeof FormSchema>;

export interface CategoryCreateProps extends UseDisclosureProps {
  defaultParent?: string;
  channelID?: string;
}

export function useCategoryCreate(props: CategoryCreateProps) {
  const { register, handleSubmit, control, setValue } = useForm<Form>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      colour: "#8577ce",
      parent: props.defaultParent,
    },
  });

  const onSubmit = handleSubmit(async (data) => {
    await handle(async () => {
      if (props.channelID) {
        // Use channel-specific API
        await channelCategoryCreate(props.channelID, data);
        // Invalidate the channel-specific category list cache
        mutate(getChannelCategoryListKey(props.channelID));
      } else {
        // Fallback to global API (will fail if no default channel set)
        await categoryCreate(data);
        mutate(getCategoryListKey());
      }
      props.onClose?.();
    });
  });

  function handleImageUpload(asset: Asset) {
    setValue("cover_image_asset_id", asset.id);
  }

  return {
    onSubmit,
    register,
    control,
    handleImageUpload,
  };
}
