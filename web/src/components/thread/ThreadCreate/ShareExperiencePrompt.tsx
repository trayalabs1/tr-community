import { useCallback } from "react";
import { useDisclosure } from "src/utils/useDisclosure";

import { TRAYA_COLORS } from "@/theme/traya-colors";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { styled } from "@/styled-system/jsx";

import { ThreadCreateModal } from "./ThreadCreateModal";

type Props = {
  channelID: string;
  title?: string;
  body?: string;
  ctaLabel?: string;
};

export function ShareExperiencePrompt({
  channelID,
  title = "Want to share something?",
  body = "Share what you're noticing at month 3 — coaches reply within a day.",
  ctaLabel = "Share my experience",
}: Props) {
  const useDisclosureProps = useDisclosure();
  const { trackEvent } = useEventTracking();

  const handleClick = useCallback(() => {
    trackEvent("community_share_experience_prompt_clicked", {
      channel_id: channelID,
    });
    useDisclosureProps.onOpen();
  }, [trackEvent, channelID, useDisclosureProps]);

  return (
    <>
      <styled.div
        width="full"
        position="relative"
        overflow="hidden"
        rounded="2xl"
        p="5"
        style={{
          background: TRAYA_COLORS.tertiary,
          border: `1px solid ${TRAYA_COLORS.secondary}`,
        }}
      >
        <styled.div
          position="absolute"
          rounded="full"
          style={{
            top: "-20px",
            right: "-20px",
            width: "120px",
            height: "120px",
            background:
              "radial-gradient(circle, rgba(255,255,255,0.6) 0%, rgba(255,255,255,0) 70%)",
            pointerEvents: "none",
          }}
        />

        <styled.div
          display="flex"
          alignItems="center"
          gap="2"
          mb="2"
          fontSize="sm"
          fontWeight="semibold"
          style={{ color: TRAYA_COLORS.primary }}
        >
          <styled.span aria-hidden="true">✨</styled.span>
          <styled.span>{title}</styled.span>
        </styled.div>

        <styled.p
          fontSize="md"
          mb="4"
          style={{ color: "#1F2937", lineHeight: 1.5 }}
        >
          {body}
        </styled.p>

        <styled.button
          type="button"
          onClick={handleClick}
          display="inline-flex"
          alignItems="center"
          gap="2"
          py="2.5"
          px="4"
          rounded="full"
          fontSize="sm"
          fontWeight="semibold"
          cursor="pointer"
          style={{
            background: TRAYA_COLORS.primary,
            color: "white",
          }}
        >
          <styled.span>{ctaLabel}</styled.span>
          <styled.span aria-hidden="true">→</styled.span>
        </styled.button>
      </styled.div>

      <ThreadCreateModal {...useDisclosureProps} channelID={channelID} />
    </>
  );
}
