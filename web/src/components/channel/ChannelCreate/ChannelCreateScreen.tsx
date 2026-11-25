"use client";

import { Button } from "@/components/ui/button";
import { FormControl } from "@/components/ui/form/FormControl";
import { FormHelperText } from "@/components/ui/form/FormHelperText";
import { FormLabel } from "@/components/ui/form/FormLabel";
import { SelectField } from "@/components/ui/form/SelectField";
import { Input, InputPrefix } from "@/components/ui/input";
import { WEB_ADDRESS } from "@/config";
import { HStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { createListCollection } from "@ark-ui/react";

import { ChannelCreateProps, useChannelCreate } from "./useChannelCreate";

export type { ChannelCreateProps };

export function ChannelCreateScreen(props: ChannelCreateProps) {
  const { register, onSubmit, control } = useChannelCreate(props);
  const hostname = new URL(WEB_ADDRESS).host;

  const visibilityCollection = createListCollection({
    items: [
      { label: "Public - Discoverable by everyone", value: "public" },
      { label: "Private - Invite-only", value: "private" },
    ],
  });

  return (
    <VStack alignItems="start" gap="4">
      <styled.p>
        Create a new channel to organize discussions around a specific topic or
        community. Channels can have their own categories and threads.
      </styled.p>
      <styled.form
        display="flex"
        flexDir="column"
        gap="4"
        w="full"
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
            The URL path for your channel (e.g., &quot;announcements&quot;,
            &quot;support&quot;)
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

        <WStack>
          <Button flexGrow="1" type="button" onClick={props.onClose}>
            Cancel
          </Button>
          <Button flexGrow="1" type="submit">
            Create Channel
          </Button>
        </WStack>
      </styled.form>
    </VStack>
  );
}
