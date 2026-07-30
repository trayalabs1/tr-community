"use client";

import { useRouter } from "next/navigation";
import { useQueryState } from "nuqs";

import {
  NotificationListResult,
  NotificationStatus,
} from "@/api/openapi-schema";
import { NotificationItem } from "@/components/notifications/item";
import { NotificationList } from "@/components/notifications/NotificationList";
import { useNotifications } from "@/components/notifications/useNotifications";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { UnreadyBanner } from "@/components/site/Unready";
import { Switch } from "@/components/ui/switch";
import { styled } from "@/styled-system/jsx";

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
  const result = useNotifications({
    initialData: props.initialData,
    status,
  });

  if (!result.ready) {
    return {
      ready: false as const,
      error: result.error,
    };
  }

  const { data, handlers } = result;

  function handleToggleStatus() {
    setStatus(
      status === NotificationStatus.unread
        ? NotificationStatus.read
        : NotificationStatus.unread,
    );
  }

  async function handleOpen(notification: NotificationItem) {
    if (notification.isRead) {
      return;
    }

    await handlers.handleMarkAs(notification.id, NotificationStatus.read);
  }

  return {
    ready: true as const,
    data,
    status,
    handlers: {
      handleToggleStatus,
      handleOpen,
      handleMarkAllAsRead: handlers.handleMarkAllAsRead,
    },
  };
}

export function NotificationScreen(props: Props) {
  const { ready, error, data, status, handlers } = useNotificationScreen(props);
  const router = useRouter();

  if (!ready) {
    return <UnreadyBanner error={error} />;
  }

  const { notifications, unreads } = data;

  const showingArchived = status === NotificationStatus.read;

  return (
    <styled.div display="flex" flexDirection="column" w="full" h="full">
      <styled.div
        position="sticky"
        top="0"
        zIndex="sticky"
        display="flex"
        alignItems="center"
        h="14"
        px="1"
        bg="bg.default"
      >
        <styled.button
          type="button"
          aria-label="Go back"
          onClick={() => router.back()}
          display="flex"
          alignItems="center"
          justifyContent="center"
          w="12"
          h="12"
          flexShrink="0"
          color="fg.default"
          bg="transparent"
          border="none"
          cursor="pointer"
        >
          <ArrowLeftIcon width="6" height="6" />
        </styled.button>

        <styled.h1
          flex="1"
          minW="0"
          fontSize="xl"
          lineHeight="[24px]"
          fontWeight="bold"
          letterSpacing="[0.2px]"
          color="fg.default"
          textAlign="center"
        >
          Notifications
        </styled.h1>

        <styled.div w="12" flexShrink="0" />
      </styled.div>

      <styled.div
        flex="1"
        display="flex"
        flexDirection="column"
        gap="4"
        w="full"
        maxW="2xl"
        mx="auto"
        p="4"
        bg="bg.notificationPage"
      >
        <styled.div
          display="flex"
          alignItems="center"
          justifyContent="space-between"
          gap="2"
          w="full"
        >
          <Switch
            size="sm"
            checked={showingArchived}
            onClick={handlers.handleToggleStatus}
          >
            Archived
          </Switch>

          {unreads > 0 && (
            <styled.button
              type="button"
              onClick={handlers.handleMarkAllAsRead}
              fontSize="sm"
              fontWeight="medium"
              color="fg.muted"
              bg="transparent"
              border="none"
              cursor="pointer"
            >
              Mark all as read
            </styled.button>
          )}
        </styled.div>

        <NotificationList
          notifications={notifications}
          onOpen={handlers.handleOpen}
        />
      </styled.div>
    </styled.div>
  );
}
