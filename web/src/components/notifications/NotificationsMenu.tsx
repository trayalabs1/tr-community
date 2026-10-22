import { useCallback, useState } from "react";
import { Portal } from "@ark-ui/react";
import Link from "next/link";

import { useSession } from "@/auth";
import { Permission } from "@/api/openapi-schema";
import { MemberAvatar } from "@/components/member/MemberBadge/MemberAvatar";
import { Button } from "@/components/ui/button";
import { ArchiveIcon } from "@/components/ui/icons/Archive";
import { SettingsIcon } from "@/components/ui/icons/Settings";
import { LinkButton } from "@/components/ui/link-button";
import * as Menu from "@/components/ui/menu";
import { Center, LStack, WStack, styled } from "@/styled-system/jsx";
import { hstack } from "@/styled-system/patterns";
import { deriveError } from "@/utils/error";
import { hasPermission } from "@/utils/permissions";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

import { NotificationsTrigger } from "./NotificationsTrigger";
import { NotificationItem } from "./item";
import { Props, useNotifications } from "./useNotifications";

export function NotificationsMenu(props: Props) {
  // For admins the list is only fetched once they first open the menu, so it
  // is not requested on every page load (admins have long notification
  // histories). It stays enabled afterwards so the SWR cache is reused and
  // mark-as-read updates keep working. Other members fetch eagerly.
  const session = useSession();
  const lazy = hasPermission(session, Permission.ADMINISTRATOR);
  const [opened, setOpened] = useState(false);
  const { ready, error, data, handlers } = useNotifications({
    ...props,
    enabled: lazy ? opened : undefined,
  });
  const { trackNotificationsClicked } = useEventTracking();

  const handleNotificationsClick = useCallback(() => {
    trackNotificationsClicked();
  }, [trackNotificationsClicked]);

  if (!ready && (error || !lazy)) {
    return (
      <NotificationsTrigger
        hideLabel
        size="md"
        variant="ghost"
        disabled
        title={deriveError(error)}
      />
    );
  }

  const unreads = ready ? data.unreads : 0;
  const notifications = ready ? data.notifications : [];

  const isEmpty = notifications.length === 0;

  return (
    <Menu.Root
      closeOnSelect={false}
      onOpenChange={(details) => {
        if (details.open) {
          setOpened(true);
        }
      }}
    >
      <Menu.Trigger cursor="pointer" position="relative" asChild onClick={handleNotificationsClick}>
        <NotificationsTrigger
          hideLabel
          size="md"
          variant="ghost"
          unread={!isEmpty}
        />
      </Menu.Trigger>

      <Portal>
        <Menu.Positioner>
          <Menu.Content minW="48" userSelect="none">
            <Menu.ItemGroup id="heading">
              <Menu.ItemGroupLabel display="flex" gap="2" alignItems="center">
                <LStack fontSize="sm">
                  <WStack>
                    <styled.p color="fg.muted">
                      Notifications ({unreads})
                    </styled.p>

                    <LinkButton
                      href="/notifications"
                      size="xs"
                      variant="outline"
                    >
                      see all
                    </LinkButton>
                  </WStack>
                </LStack>
              </Menu.ItemGroupLabel>

              <Menu.Separator />

              {!ready ? (
                <Center w="full" py="4" color="fg.muted" fontSize="xs">
                  Loading…
                </Center>
              ) : isEmpty ? (
                <Center w="full" py="4" color="fg.muted" fontSize="xs">
                  You&apos;re all caught up!
                </Center>
              ) : (
                notifications.map((notification) => (
                  <Menu.Item
                    key={notification.id}
                    value={notification.id}
                    height="auto"
                    py="1"
                  >
                    <WStack className="notification-menu__row" minW="0" gap="1">
                      <Link
                        className={hstack({
                          w: "full",
                          minW: "0",
                          justify: "space-between",
                        })}
                        href={notification.url}
                        onClick={() =>
                          handlers.handleMarkAs(notification.id, "read")
                        }
                      >
                        <NotificationAvatar notification={notification} />
                        <LStack gap="0" minW="0">
                          <styled.span
                            fontWeight="bold"
                            textWrap="nowrap"
                            textOverflow="ellipsis"
                            overflow="hidden"
                            maxW="full"
                          >
                            {notification.source?.handle ?? "System"}
                          </styled.span>
                          <styled.span fontWeight="normal">
                            {notification.description}
                          </styled.span>
                        </LStack>
                      </Link>

                      <Button
                        variant="ghost"
                        size="sm"
                        title="Mark as read"
                        onClick={() =>
                          handlers.handleMarkAs(notification.id, "read")
                        }
                      >
                        <ArchiveIcon />
                      </Button>
                    </WStack>
                  </Menu.Item>
                ))
              )}
            </Menu.ItemGroup>
          </Menu.Content>
        </Menu.Positioner>
      </Portal>
    </Menu.Root>
  );
}

export function NotificationAvatar(props: { notification: NotificationItem }) {
  if (props.notification.source) {
    return <MemberAvatar profile={props.notification.source} size="sm" />;
  }

  return <SettingsIcon />;
}
