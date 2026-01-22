"use client";

import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { mutate } from "swr";
import { createListCollection } from "@ark-ui/react";
import { useEffect, useState } from "react";

import { Account, Channel, Asset } from "@/api/openapi-schema";
import { Button } from "@/components/ui/button";
import { FormControl } from "@/components/ui/form/FormControl";
import { FormHelperText } from "@/components/ui/form/FormHelperText";
import { FormLabel } from "@/components/ui/form/FormLabel";
import { SelectField } from "@/components/ui/form/SelectField";
import { Input, InputPrefix } from "@/components/ui/input";
import { Heading } from "@/components/ui/heading";
import { WEB_ADDRESS } from "@/config";
import { HStack, LStack, WStack, styled } from "@/styled-system/jsx";
import {
  channelUpdate,
  getChannelListKey,
  getChannelGetKey,
} from "@/api/openapi-client/channels";
import { handle } from "@/api/client";
import { useChannelPermissions } from "@/lib/channel/permissions";
import { AssetUploadEditor } from "@/components/asset/AssetUploadEditor/AssetUploadEditor";

import { revalidateChannels } from "@/app/(dashboard)/channels/actions";
import { MembersSection } from "./MembersSection";
import { CategoriesSection } from "./CategoriesSection";

type Props = {
  session?: Account;
  channel: Channel;
};

const FormSchema = z.object({
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
  visibility: z.enum(["public", "private"]),
});
type Form = z.infer<typeof FormSchema>;

export function ChannelSettingsScreen(props: Props) {
  const router = useRouter();
  const hostname = new URL(WEB_ADDRESS).host;
  const permissions = useChannelPermissions(props.channel.id);
  const [selectedIcon, setSelectedIcon] = useState<Asset | undefined>(props.channel.icon);

  useEffect(() => {
    if (permissions.role !== null && !permissions.canManageChannel) {
      router.push(`/channels/${props.channel.id}`);
    }
  }, [permissions.canManageChannel, permissions.role, props.channel.id, router]);

  useEffect(() => {
    setSelectedIcon(props.channel.icon);
  }, [props.channel.id]);

  const { register, handleSubmit, control, formState } = useForm<Form>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      name: props.channel.name,
      slug: props.channel.slug,
      description: props.channel.description,
      visibility: props.channel.visibility as "public" | "private",
    },
  });

  const visibilityCollection = createListCollection({
    items: [
      { label: "Public - Discoverable by everyone", value: "public" },
      { label: "Private - Invite-only", value: "private" },
    ],
  });

  const onSubmit = handleSubmit(async (data) => {
    await handle(async () => {
      const updateData = {
        ...data,
        icon: selectedIcon?.id || null,
      };

      await channelUpdate(props.channel.id, updateData);

      mutate(getChannelListKey());
      mutate(getChannelGetKey(props.channel.id));
      await revalidateChannels(props.channel.id);
      router.refresh();
      router.push(`/channels/${props.channel.id}`);
    });
  });

  if (permissions.role === null || !permissions.canManageChannel) {
    return null;
  }

  return (
    <LStack gap="6" p="4">
      <HStack gap="2" alignItems="center">
        <Link href={`/channels/${props.channel.id}`}>
          <styled.span
            color="fg.muted"
            fontSize="sm"
            _hover={{ textDecoration: "underline" }}
          >
            {props.channel.name}
          </styled.span>
        </Link>
        <styled.span color="fg.muted">/</styled.span>
        <styled.span fontSize="sm">Settings</styled.span>
      </HStack>

      <Heading as="h1" size="2xl">
        Channel Settings
      </Heading>

      <styled.form
        display="flex"
        flexDir="column"
        gap="4"
        w="full"
        maxW="2xl"
        onSubmit={onSubmit}
      >
        <FormControl>
          <FormLabel>Name</FormLabel>
          <Input {...register("name")} type="text" />
          <FormHelperText>The display name of your channel</FormHelperText>
        </FormControl>

        <FormControl>
          <FormLabel>URL Slug</FormLabel>
          <HStack gap="0" alignItems="stretch" flex="1">
            <InputPrefix
              display={{
                base: "none",
                sm: "flex",
              }}
            >
              {hostname}/channels/
            </InputPrefix>
            <Input
              {...register("slug")}
              type="text"
              flex="1"
              borderTopLeftRadius={{
                base: "sm",
                sm: "none",
              }}
              borderBottomLeftRadius={{
                base: "sm",
                sm: "none",
              }}
            />
          </HStack>
          <FormHelperText>
            The URL path for your channel
          </FormHelperText>
        </FormControl>

        <FormControl>
          <FormLabel>Description</FormLabel>
          <Input {...register("description")} type="text" />
          <FormHelperText>Describe the purpose of this channel</FormHelperText>
        </FormControl>

        <FormControl>
          <FormLabel>Visibility</FormLabel>
          <SelectField
            name="visibility"
            control={control}
            collection={visibilityCollection}
            placeholder="Select visibility"
          />
          <FormHelperText>
            Public channels are discoverable, private channels are invite-only
          </FormHelperText>
        </FormControl>

        <FormControl>
          <FormLabel>Channel Icon</FormLabel>
          <AssetUploadEditor
            width={200}
            height={200}
            onUpload={setSelectedIcon}
            value={selectedIcon}
          />
          <FormHelperText>Upload a custom icon for your channel</FormHelperText>
        </FormControl>

        <WStack>
          <Button
            flexGrow="1"
            type="button"
            variant="ghost"
            onClick={() => router.back()}
          >
            Cancel
          </Button>
          <Button
            flexGrow="1"
            type="submit"
            loading={formState.isSubmitting}
          >
            Save Changes
          </Button>
        </WStack>
      </styled.form>

      <styled.div
        style={{ border: "1px solid var(--colors-border-default)" }}
        borderRadius="md"
        p="6"
        maxW="2xl"
      >
        <MembersSection channelID={props.channel.id} />
      </styled.div>

      <styled.div
        style={{ border: "1px solid var(--colors-border-default)" }}
        borderRadius="md"
        p="6"
        maxW="2xl"
      >
        <CategoriesSection channelID={props.channel.id} />
      </styled.div>
    </LStack>
  );
}
