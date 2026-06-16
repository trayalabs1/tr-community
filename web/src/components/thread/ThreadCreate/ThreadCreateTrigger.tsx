import { useCallback, useState } from "react";
import { useDisclosure } from "src/utils/useDisclosure";

import { useSession } from "@/auth";
import { ButtonProps } from "@/components/ui/button";
import { PromptNudge } from "@/components/feed/PromptNudge/PromptNudge";
import { PromptItem } from "@/components/feed/PromptNudge/prompts";
import { EditIcon } from "@/components/ui/icons/Edit";
import { TRAYA_COLORS } from "@/theme/traya-colors";
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
  const firstName = session?.name?.split(" ")[0];
  const placeholder = firstName
    ? `Type your questions ${firstName}`
    : "Type your questions & doubts here...";
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
        gap="3"
        w="full"
        py="3"
        px="4"
        rounded="full"
        cursor="pointer"
        style={{
          background: "white",
          border: `1.5px solid ${TRAYA_COLORS.primary}`,
        }}
        {...props}
      >
        <styled.div
          display="flex"
          alignItems="center"
          justifyContent="center"
          flexShrink="0"
          rounded="full"
          style={{
            width: "32px",
            height: "32px",
            backgroundColor: TRAYA_COLORS.tertiary,
          }}
        >
          <EditIcon
            style={{
              width: "16px",
              height: "16px",
              color: TRAYA_COLORS.primary,
            }}
          />
        </styled.div>
        <styled.span
          fontSize="sm"
          style={{ color: TRAYA_COLORS.neutral.textMuted }}
        >
          {placeholder}
        </styled.span>
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
        initialText={picked?.prompt.text}
        tag={picked?.prompt.tag}
        tagIndex={picked ? picked.index + 1 : undefined}
      />
    </LStack>
  );
}
