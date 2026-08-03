import { useCallback, useState } from "react";
import { useDisclosure } from "src/utils/useDisclosure";

import { useSession } from "@/auth";
import { ButtonProps } from "@/components/ui/button";
import { PromptNudge } from "@/components/feed/PromptNudge/PromptNudge";
import { PromptItem } from "@/components/feed/PromptNudge/prompts";
import { Send } from "lucide-react";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { LStack, styled } from "@/styled-system/jsx";

import { PromptComposeSheet } from "./PromptComposeSheet";
import { ThreadCreateModal } from "./ThreadCreateModal";

type Props = Omit<ButtonProps, "onClick"> & {
  channelID: string;
  channelName?: string;
  promptNudges?: PromptItem[];
};

export function ThreadCreateTrigger({
  channelID,
  channelName,
  promptNudges,
  ...props
}: Props) {
  const useDisclosureProps = useDisclosure();
  const promptSheet = useDisclosure();
  const { trackEvent } = useEventTracking();
  const session = useSession();
  const placeholder = "What's in your mind?";
  const [picked, setPicked] = useState<
    { prompt: PromptItem; index: number } | undefined
  >(undefined);

  const handleCreatePostClick = useCallback(() => {
    trackEvent("community_create_post_clicked", {
      channel_id: channelID,
      source: "post_nudge_start",
    });
    useDisclosureProps.onOpen();
  }, [trackEvent, channelID, useDisclosureProps]);

  const handlePromptPick = useCallback(
    (prompt: PromptItem, index: number) => {
      trackEvent("community_create_post_clicked", {
        channel_id: channelID,
        source: "post_nudge_select",
      });
      setPicked({ prompt, index });
      promptSheet.onOpen();
    },
    [trackEvent, channelID, promptSheet],
  );

  return (
    <LStack gap="2">
      <styled.button
        type="button"
        onClick={handleCreatePostClick}
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        gap="3"
        w="full"
        cursor="pointer"
        style={{
          background: "white",
          border: "1px solid white",
          borderRadius: "14px",
          padding: "14px 16px",
          boxShadow: "0px 4px 6px rgba(0,0,0,0.12)",
        }}
        {...props}
      >
        <styled.span
          textAlign="left"
          style={{
            color: "#999999",
            fontSize: "14px",
            fontWeight: 500,
            lineHeight: "20px",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {placeholder}
        </styled.span>
        <Send
          size={20}
          style={{
            flexShrink: 0,
            color: "var(--colors-fg-default)",
          }}
        />
      </styled.button>

      {promptNudges && promptNudges.length > 0 && (
        <PromptNudge prompts={promptNudges} onPick={handlePromptPick} />
      )}

      <ThreadCreateModal {...useDisclosureProps} channelID={channelID} />

      <PromptComposeSheet
        {...promptSheet}
        channelID={channelID}
        channelName={channelName}
        session={session}
        initialText={picked?.prompt.placeholder ?? picked?.prompt.text}
        tag={picked?.prompt.tag}
        tagIndex={picked ? picked.index + 1 : undefined}
      />
    </LStack>
  );
}
