"use client";

import { ArrowLeftIcon } from "lucide-react";
import { useRouter } from "next/navigation";
import { useQueryState } from "nuqs";

import {
  NotificationListResult,
  NotificationStatus,
} from "@/api/openapi-schema";
import { NotificationCardList } from "@/components/notifications/NotificationCardList";
import { useNotifications } from "@/components/notifications/useNotifications";
import { UnreadyBanner } from "@/components/site/Unready";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { HStack, LStack, WStack, styled } from "@/styled-system/jsx";

type Props = {
  initialData: NotificationListResult;
};

export function useNotificationScreen(props: Props) {
  const [status, setStatus] = useQueryState<NotificationStatus>("status", {
    defaultValue: "unread",
    parse(v: string) {
      switch (v) {
        case "read":
          return NotificationStatus.read;
        default:
          return NotificationStatus.unread;
      }
    },
  });
  const { ready, data, error, handlers } = useNotifications({
    initialData: props.initialData,
    status,
  });
  if (!ready) {
    return {
      ready: false as const,
      error,
    };
  }

  function handleToggleStatus() {
    setStatus(
      status === NotificationStatus.unread
        ? NotificationStatus.read
        : NotificationStatus.unread,
    );
  }

  return {
    ready: true as const,
    data,
    status,
    handlers: {
      handleToggleStatus,
      handleMarkAs: handlers.handleMarkAs,
      handleMarkAllAsRead: handlers.handleMarkAllAsRead,
    },
  };
}

export function NotificationScreen(props: Props) {
  const { ready, error, data, status, handlers } = useNotificationScreen(props);
  if (!ready) {
    return <UnreadyBanner error={error} />;
  }

  const { notifications } = data;

  const showingArchived = status === NotificationStatus.read;

  const hasUnreadNotifications = data.unreads > 0;

  const router = useRouter();

  return (
    <LStack>
      <WStack justifyContent="space-between" alignItems="flex-start">
        <LStack>
          <HStack
            gap="0"
            p="2"
            alignItems="center"
            width="full"
            display={{ base: "flex", md: "none" }}
            borderBottomWidth="thin"
            borderBottomColor="border.default"
            style={{
              position: "sticky",
              top: 0,
              zIndex: 10,
              backgroundColor: "rgba(255, 255, 255, 0.95)",
              backdropFilter: "blur(12px)",
              flexShrink: 0,
              borderBottomColor: "rgba(0, 0, 0, 0.05)",
            }}
          >
            <styled.button
              onClick={() => router.back()}
              p="2"
              style={{
                marginLeft: "-0.5rem",
                background: "transparent",
                border: "none",
                cursor: "pointer",
                borderRadius: "0.75rem",
                transition: "background-color 0.2s",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.backgroundColor =
                  "rgba(0, 0, 0, 0.05)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.backgroundColor =
                  "transparent";
              }}
            >
              <ArrowLeftIcon width="16" height="16" />
            </styled.button>
            <styled.h1 fontSize="md" fontWeight="bold">
              Notifications
            </styled.h1>
          </HStack>
          <HStack alignItems="center" gap="2" justifyContent="space-between" width="full">
            <Switch
              size="sm"
              checked={showingArchived}
              onClick={handlers.handleToggleStatus}
            >
              Archived
            </Switch>
            { hasUnreadNotifications && (
              <styled.span fontSize="sm" color="fg.muted" cursor="pointer" onClick={handlers.handleMarkAllAsRead}>
                Mark all as read
              </styled.span>
            )}
          </HStack>
        </LStack>

      </WStack>

      <NotificationCardList
        notifications={notifications}
        onMove={handlers.handleMarkAs}
      />
    </LStack>
  );
}
