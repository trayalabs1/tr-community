import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { mutate } from "swr";
import { z } from "zod";

import { handle } from "@/api/client";
import {
  channelCreate,
  getChannelListKey,
} from "@/api/openapi-client/channels";
import { UseDisclosureProps } from "@/utils/useDisclosure";

export const FormSchema = z.object({
  name: z
    .string()
    .min(1, "Please enter a name for the channel.")
    .max(100, "Name must be 100 characters or less."),
  slug: z
    .string()
    .min(1, "Please enter a URL slug for the channel.")
    .max(100, "Slug must be 100 characters or less.")
    .regex(/^[a-z0-9-]+$/, "Slug can only contain lowercase letters, numbers, and hyphens."),
  description: z
    .string()
    .min(1, "Please enter a description.")
    .max(500, "Description must be 500 characters or less."),
  visibility: z.enum(["public", "private"]).default("public"),
});
export type Form = z.infer<typeof FormSchema>;

export interface ChannelCreateProps extends UseDisclosureProps {}

export function useChannelCreate(props: ChannelCreateProps) {
  const router = useRouter();
  const { register, handleSubmit, control } = useForm<Form>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      visibility: "public",
    },
  });

  const onSubmit = handleSubmit(async (data) => {
    await handle(async () => {
      const channel = await channelCreate(data);
      props.onClose?.();
      mutate(getChannelListKey());
      router.refresh(); // Invalidate Next.js router cache for server components
      router.push(`/channels/${channel.id}`);
    });
  });

  return {
    onSubmit,
    register,
    control,
  };
}
