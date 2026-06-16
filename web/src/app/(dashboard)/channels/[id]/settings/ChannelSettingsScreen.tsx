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
import { IconButton } from "@/components/ui/icon-button";
import { AddIcon } from "@/components/ui/icons/Add";
import { DeleteIcon } from "@/components/ui/icons/Delete";
import {
  DEFAULT_PROMPT_ICON,
  PROMPT_ICON_KEYS,
  parsePromptNudges,
  type PromptItem,
} from "@/components/feed/PromptNudge/prompts";

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
  const [prompts, setPrompts] = useState<PromptItem[]>(() =>
    parsePromptNudges(props.channel.meta),
  );

  useEffect(() => {
    setPrompts(parsePromptNudges(props.channel.meta));
  }, [props.channel.id]);

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

  function addPrompt() {
    setPrompts((prev) => [...prev, { icon: DEFAULT_PROMPT_ICON, text: "" }]);
  }

  function updatePrompt(index: number, patch: Partial<PromptItem>) {
    setPrompts((prev) =>
      prev.map((p, i) => (i === index ? { ...p, ...patch } : p)),
    );
  }

  function removePrompt(index: number) {
    setPrompts((prev) => prev.filter((_, i) => i !== index));
  }

  const onSubmit = handleSubmit(async (data) => {
    await handle(async () => {
      const promptNudges = prompts
        .map((p) => {
          const tag = p.tag?.trim();
          return {
            icon: p.icon,
            text: p.text.trim(),
            ...(tag ? { tag } : {}),
          };
        })
        .filter((p) => p.text !== "");

      const updateData = {
        ...data,
        icon: selectedIcon?.id || null,
        meta: {
          ...(props.channel.meta ?? {}),
          prompt_nudges: promptNudges,
        },
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

        <FormControl>
          <FormLabel>Posting prompts</FormLabel>
          <LStack gap="2">
            {prompts.map((prompt, index) => (
              <HStack key={index} gap="2" width="full" alignItems="center">
                <styled.select
                  value={prompt.icon}
                  onChange={(e) => updatePrompt(index, { icon: e.target.value })}
                  rounded="sm"
                  px="2"
                  py="1.5"
                  fontSize="sm"
                  bg="bg.default"
                  style={{ border: "1px solid var(--colors-border-default)" }}
                >
                  {PROMPT_ICON_KEYS.map((key) => (
                    <option key={key} value={key}>
                      {key}
                    </option>
                  ))}
                </styled.select>
                <Input
                  flex="1"
                  type="text"
                  value={prompt.text}
                  placeholder="Prompt text shown to members"
                  onChange={(e) => updatePrompt(index, { text: e.target.value })}
                />
                <Input
                  width="32"
                  type="text"
                  value={prompt.tag ?? ""}
                  placeholder="Tag (optional)"
                  onChange={(e) => updatePrompt(index, { tag: e.target.value })}
                />
                <IconButton
                  type="button"
                  variant="ghost"
                  size="sm"
                  aria-label="Remove prompt"
                  onClick={() => removePrompt(index)}
                >
                  <DeleteIcon />
                </IconButton>
              </HStack>
            ))}
            <Button type="button" variant="outline" size="sm" onClick={addPrompt}>
              <AddIcon />
              Add prompt
            </Button>
          </LStack>
          <FormHelperText>
            Suggestions shown in the “Not sure what to post?” pill. Tapping one
            opens the composer with the text as a placeholder; the optional tag
            appears as a chip (e.g. “Just Started”).
          </FormHelperText>
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
