"use client";

import Link from "next/link";
import { useCallback } from "react";
import { NotificationIcon } from "@/components/ui/icons/Notification";
import { Box, styled } from "@/styled-system/jsx";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

type Props = {
  hasUnread?: boolean;
};

export function NotificationButton({ hasUnread = false }: Props) {
  const { trackNotificationsClicked } = useEventTracking();

  const handleClick = useCallback(() => {
    trackNotificationsClicked();
  }, [trackNotificationsClicked]);

  return (
    <Link href="/notifications" onClick={handleClick}>
      <styled.div position="relative" display="inline-flex">
        <styled.button
          display="flex"
          alignItems="center"
          justifyContent="center"
          w="10"
          h="10"
          rounded="md"
          style={{
            background: "transparent",
            border: "none",
            cursor: "pointer",
            color: "var(--colors-fg-default)",
            padding: "0",
            transition: "color 0.2s ease-in-out",
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.color = "#4a9d6f";
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.color = "var(--colors-fg-default)";
          }}
          aria-label="Notifications"
        >
          <NotificationIcon width="5" height="5" />
        </styled.button>
        {hasUnread && (
          <Box
            position="absolute"
            top="1"
            right="1"
            bgColor="fg.destructive/60"
            borderRadius="full"
            w="2"
            h="2"
          />
        )}
      </styled.div>
    </Link>
  );
}
