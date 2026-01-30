import { useCallback } from "react";
import { useDisclosure } from "src/utils/useDisclosure";

import { ButtonProps } from "@/components/ui/button";
import { Button } from "@/components/ui/button";
import { AddIcon } from "@/components/ui/icons/Add";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

import { ThreadCreateModal } from "./ThreadCreateModal";

type Props = ButtonProps & {
  channelID: string;
};

export function ThreadCreateTrigger({ channelID, ...props }: Props) {
  const useDisclosureProps = useDisclosure();
  const { trackEvent } = useEventTracking();

  const handleCreatePostClick = useCallback(() => {
    trackEvent("community_create_post_clicked", { channel_id: channelID });
    useDisclosureProps.onOpen();
  }, [trackEvent, channelID, useDisclosureProps]);

  return (
    <>
      <Button
        size="sm"
        onClick={handleCreatePostClick}
        style={{
          background: TRAYA_COLORS.primary,
          color: "white",
          borderRadius: "0.875rem",
          padding: "0.5rem 0.75rem",
          fontWeight: "500",
          fontSize: "14px",
          transition: "all 0.2s ease-in-out",
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.opacity = "0.9";
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.opacity = "1";
        }}
        {...props}
      >
        <AddIcon width="4" height="4" />
        Create Post
      </Button>

      <ThreadCreateModal
        {...useDisclosureProps}
        channelID={channelID}
      />
    </>
  );
}
