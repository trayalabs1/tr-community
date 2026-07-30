import Link from "next/link";

import { MemberAvatar } from "@/components/member/MemberBadge/MemberAvatar";
import { styled } from "@/styled-system/jsx";
import { timestamp } from "@/utils/date";

import { NotificationItem } from "./item";

type Props = {
  notification: NotificationItem;
  onOpen: (notification: NotificationItem) => void;
};

export function NotificationRow({ notification, onOpen }: Props) {
  const { isRead, source, title, description, url, createdAt } = notification;

  return (
    <styled.li listStyleType="none" w="full">
      <Link
        href={url}
        onClick={() => onOpen(notification)}
        style={{ textDecoration: "none", display: "block" }}
      >
        <styled.div
          display="flex"
          alignItems="center"
          gap="3"
          w="full"
          px="3"
          py="3"
          borderRadius="2xl"
          borderWidth={isRead ? "none" : "thin"}
          borderStyle="solid"
          borderColor={isRead ? "transparent" : "border.default"}
          bg={isRead ? "bg.notificationRead" : "bg.notificationUnread"}
        >
          {source ? (
            <MemberAvatar profile={source} size="md" />
          ) : (
            <styled.div
              flexShrink="0"
              w="9"
              h="9"
              borderRadius="full"
              bg="bg.muted"
            />
          )}

          <styled.div
            display="flex"
            flexDirection="column"
            gap="0.5"
            flex="1"
            minW="0"
          >
            <styled.p
              fontSize="sm"
              lineHeight="[20px]"
              fontWeight="bold"
              color={isRead ? "fg.muted" : "fg.default"}
              truncate
            >
              {title}
            </styled.p>

            <styled.p
              fontSize="sm"
              lineHeight="[20px]"
              fontWeight="medium"
              color={isRead ? "fg.muted" : "fg.subtle"}
              truncate
            >
              {description}
            </styled.p>
          </styled.div>

          <styled.span
            flexShrink="0"
            fontSize="xs"
            lineHeight="[16px]"
            fontWeight="medium"
            letterSpacing="[0.24px]"
            color={isRead ? "fg.muted" : "fg.default"}
          >
            {timestamp(createdAt)}
          </styled.span>
        </styled.div>
      </Link>
    </styled.li>
  );
}
