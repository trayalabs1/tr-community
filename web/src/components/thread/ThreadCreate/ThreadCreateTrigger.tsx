import { useCallback } from "react";
import { useDisclosure } from "src/utils/useDisclosure";

import { useSession } from "@/auth";
import { ButtonProps } from "@/components/ui/button";
import { EditIcon } from "@/components/ui/icons/Edit";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { styled } from "@/styled-system/jsx";

import { ThreadCreateModal } from "./ThreadCreateModal";

type Props = Omit<ButtonProps, "onClick"> & {
  channelID: string;
};

export function ThreadCreateTrigger({ channelID, ...props }: Props) {
  const useDisclosureProps = useDisclosure();
  const { trackEvent } = useEventTracking();
  const session = useSession();
  const firstName = session?.name?.split(" ")[0];
  const placeholder = firstName
    ? `Type your questions ${firstName}`
    : "Type your questions & doubts here...";

  const handleCreatePostClick = useCallback(() => {
    trackEvent("community_create_post_clicked", { channel_id: channelID });
    useDisclosureProps.onOpen();
  }, [trackEvent, channelID, useDisclosureProps]);

  return (
    <>
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

      <ThreadCreateModal
        {...useDisclosureProps}
        channelID={channelID}
      />
    </>
  );
}
