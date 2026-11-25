import { styled } from "@/styled-system/jsx";

import { NotificationItem } from "./item";
import { NotificationRow } from "./NotificationRow";

type Props = {
  notifications: NotificationItem[];
  onOpen: (notification: NotificationItem) => void;
};

export function NotificationList({ notifications, onOpen }: Props) {
  if (notifications.length === 0) {
    return (
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="center"
        w="full"
        py="16"
      >
        <styled.p fontSize="sm" color="fg.muted">
          No notifications.
        </styled.p>
      </styled.div>
    );
  }

  return (
    <styled.ol display="flex" flexDirection="column" gap="2" w="full" p="0" m="0">
      {notifications.map((n) => (
        <NotificationRow key={n.id} notification={n} onOpen={onOpen} />
      ))}
    </styled.ol>
  );
}
