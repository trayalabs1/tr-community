import Link from "next/link";

import { ThreadReference } from "src/api/openapi-schema";
import { CollectionMenu } from "src/components/content/CollectionMenu/CollectionMenu";
import { useSession } from "src/auth";

import { styled } from "@/styled-system/jsx";
import { formatDateTime } from "@/utils/date";
import { getAvatarColor } from "@/utils/avatar-colors";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { Badge } from "../ui/badge";
import { SharedFromIcon } from "../ui/icons/Arrow";
import {
  DiscussionIcon,
  DiscussionParticipatingIcon,
} from "../ui/icons/Discussion";

import { ThreadMenu } from "../thread/ThreadMenu/ThreadMenu";
import { LikeButton } from "./LikeButton/LikeButton";
import { ProfileHoverTooltip } from "./ProfileHoverTooltip";

type Props = {
  thread: ThreadReference;
  permalink: string;
  firstImageUrl: string | null;
  channelID?: string;
};

export function SharedThreadCard({
  thread,
  permalink,
  firstImageUrl,
  channelID,
}: Props) {
  const session = useSession();

  const sharedBy = thread.shared_by;
  const sharedFrom = thread.shared_from;

  const title = thread.title || thread.link?.title || "Untitled post";

  const hasReplied = thread.reply_status.replied > 0;
  const replyCount = thread.reply_status.replies;
  const replyCountLabel =
    replyCount === 1 ? `1 reply` : `${replyCount} replies`;
  const replyStatusLabel = hasReplied
    ? `${replyCountLabel} (including you!)`
    : replyCountLabel;

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: `1px solid ${TRAYA_COLORS.border.default}`,
        boxShadow: TRAYA_COLORS.shadow.subtle,
        transition: "all 0.2s ease-in-out",
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow =
          TRAYA_COLORS.shadow.medium;
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow =
          TRAYA_COLORS.shadow.subtle;
      }}
    >
      {/* Outer header: the sharer (admin) */}
      <styled.div
        display="flex"
        alignItems="center"
        gap="3"
        px="4"
        pt="4"
        pb="2"
      >
        {sharedBy && (
          <ProfileHoverTooltip profile={sharedBy}>
            <Link
              href={`/m/${sharedBy.handle}`}
              style={{ textDecoration: "none" }}
            >
              <styled.button
                w="10"
                h="10"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                fontSize="sm"
                fontWeight="semibold"
                style={{
                  backgroundColor: getAvatarColor(sharedBy.handle),
                  color: "white",
                  border: "none",
                  cursor: "pointer",
                  flexShrink: 0,
                }}
                aria-label={`View profile of @${sharedBy.handle}`}
              >
                {sharedBy.handle.charAt(0).toUpperCase()}
              </styled.button>
            </Link>
          </ProfileHoverTooltip>
        )}

        <styled.div
          display="flex"
          flexDir="column"
          gap="0.5"
          flex="1"
          minW="0"
        >
          <styled.div display="flex" alignItems="center" gap="2" flexWrap="wrap">
            {sharedBy && (
              <Link
                href={`/m/${sharedBy.handle}`}
                style={{ textDecoration: "none" }}
              >
                <styled.p
                  fontSize="sm"
                  fontWeight="semibold"
                  color="fg.default"
                  style={{
                    margin: "0",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    cursor: "pointer",
                  }}
                >
                  @{sharedBy.handle}
                </styled.p>
              </Link>
            )}

            {/* Only administrators can share a thread into a channel, so the
                sharer is definitionally an admin whenever `shared_by` is set. */}
            <Badge
              size="sm"
              aria-label="Administrator"
              style={{
                backgroundColor: TRAYA_COLORS.secondary,
                color: TRAYA_COLORS.primary,
                border: "none",
              }}
            >
              ADMIN
            </Badge>
          </styled.div>

          <styled.p
            fontSize="xs"
            color="fg.muted"
            style={{ margin: "0" }}
          >
            {formatDateTime(thread.createdAt)}
          </styled.p>
        </styled.div>

        {session && (
          <styled.div flexShrink="0">
            <ThreadMenu thread={thread} channelID={channelID} />
          </styled.div>
        )}
      </styled.div>

      {/* Subtitle: the admin's editorial text, rendered as the prominent body */}
      <styled.div px="4" pt="1" pb="3">
        {thread.subtitle && (
          <styled.p
            fontSize="sm"
            fontWeight="medium"
            color="fg.default"
            style={{ margin: "0 0 12px 0", lineHeight: "1.6" }}
          >
            {thread.subtitle}
          </styled.p>
        )}

        {/* Nested card: the original thread being shared */}
        <Link href={permalink} style={{ textDecoration: "none" }}>
          <styled.div
            display="flex"
            flexDir="column"
            borderRadius="xl"
            overflow="hidden"
            cursor="pointer"
            style={{
              border: `1px solid ${TRAYA_COLORS.secondary}`,
              transition: "border-color 0.2s ease-in-out",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.borderColor =
                TRAYA_COLORS.primary;
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.borderColor =
                TRAYA_COLORS.secondary;
            }}
          >
            {sharedFrom && (
              <styled.div
                display="flex"
                alignItems="center"
                gap="1.5"
                px="3"
                py="2"
                style={{
                  backgroundColor: TRAYA_COLORS.tertiary,
                  borderBottom: `1px solid ${TRAYA_COLORS.secondary}`,
                }}
              >
                <SharedFromIcon width="3.5" height="3.5" style={{ color: TRAYA_COLORS.primary }} />
                <styled.span
                  fontSize="xs"
                  fontWeight="medium"
                  style={{ color: TRAYA_COLORS.primary }}
                >
                  From
                </styled.span>
                <styled.span
                  fontSize="xs"
                  fontWeight="bold"
                  style={{ color: TRAYA_COLORS.primary }}
                >
                  {sharedFrom.name}
                </styled.span>
              </styled.div>
            )}

            <styled.div display="flex" flexDir="column" gap="2" p="3">
              <styled.div display="flex" alignItems="center" gap="2">
                <styled.div
                  w="7"
                  h="7"
                  rounded="full"
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  fontSize="xs"
                  fontWeight="semibold"
                  style={{
                    backgroundColor: getAvatarColor(thread.author.handle),
                    color: "white",
                    flexShrink: 0,
                  }}
                >
                  {thread.author.handle.charAt(0).toUpperCase()}
                </styled.div>
                <styled.div display="flex" flexDir="column" gap="0" minW="0">
                  <styled.span
                    fontSize="xs"
                    fontWeight="semibold"
                    color="fg.default"
                    style={{
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    @{thread.author.handle}
                  </styled.span>
                  <styled.span fontSize="2xs" color="fg.muted">
                    {formatDateTime(thread.createdAt)}
                  </styled.span>
                </styled.div>
              </styled.div>

              <styled.p
                fontSize="sm"
                fontWeight="semibold"
                color="fg.default"
                style={{ margin: "0", lineHeight: "1.4" }}
              >
                {title}
              </styled.p>

              {thread.description && (
                <styled.p
                  fontSize="xs"
                  color="fg.muted"
                  style={{
                    margin: "0",
                    lineHeight: "1.5",
                    overflow: "hidden",
                    display: "-webkit-box",
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: "vertical",
                    textOverflow: "ellipsis",
                  }}
                >
                  {thread.description}
                </styled.p>
              )}

              {firstImageUrl && (
                <styled.div
                  mt="1"
                  borderRadius="lg"
                  overflow="hidden"
                  style={{ height: "140px", width: "100%" }}
                >
                  <styled.img
                    src={firstImageUrl}
                    alt=""
                    style={{
                      width: "100%",
                      height: "100%",
                      objectFit: "cover",
                    }}
                  />
                </styled.div>
              )}
            </styled.div>
          </styled.div>
        </Link>
      </styled.div>

      {/* Outer footer: likes, replies, bookmark */}
      {session && (
        <styled.div
          display="flex"
          alignItems="center"
          gap="4"
          p="4"
          style={{
            borderTop: "1px solid var(--colors-border-default)/50",
          }}
        >
          <LikeButton thread={thread} showCount />

          <Link href={permalink} title={replyStatusLabel}>
            <styled.div
              display="flex"
              alignItems="center"
              gap="1.5"
              color="fg.muted"
              style={{ fontSize: "14px" }}
            >
              {hasReplied ? (
                <DiscussionParticipatingIcon width="5" />
              ) : (
                <DiscussionIcon width="5" />
              )}
              <styled.span fontSize="sm" fontWeight="medium">
                {replyCount}
              </styled.span>
            </styled.div>
          </Link>

          <styled.div style={{ marginLeft: "auto" }}>
            <CollectionMenu account={session} thread={thread} />
          </styled.div>
        </styled.div>
      )}
    </styled.div>
  );
}
