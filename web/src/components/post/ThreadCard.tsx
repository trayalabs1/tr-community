import Link from "next/link";
import { memo } from "react";

import {
  Permission,
  ThreadReference,
  Visibility,
} from "src/api/openapi-schema";
import { useSession } from "src/auth";
import { CollectionMenu } from "src/components/content/CollectionMenu/CollectionMenu";

import { styled } from "@/styled-system/jsx";
import { timestamp } from "@/utils/date";
import { hasPermission } from "@/utils/permissions";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { CategoryBadge } from "../category/CategoryBadge";
import { PostReviewBadge } from "../thread/PostReviewBadge";
import { ThreadMenu } from "../thread/ThreadMenu/ThreadMenu";
import {
  DiscussionIcon,
  DiscussionParticipatingIcon,
} from "../ui/icons/Discussion";

import { LikeButton } from "./LikeButton/LikeButton";
import { useThreadCardModeration } from "./useThreadCardModeration";
import { ProfileHoverTooltip } from "./ProfileHoverTooltip";

type Props = {
  thread: ThreadReference;
  hideCategoryBadge?: boolean;
  channelID?: string;
};

export const ThreadReferenceCard = memo(
  ({ thread, hideCategoryBadge = false, channelID }: Props) => {
    const session = useSession();
    const isDraft = thread.visibility === Visibility.draft;
    const permalink = isDraft
      ? `/new?id=${thread.id}`
      : channelID
        ? `/channels/${channelID}/threads/${thread.slug}`
        : `/t/${thread.slug}`;
    const isModerator = hasPermission(
      session,
      Permission.MANAGE_POSTS,
      Permission.ADMINISTRATOR,
    );

    const { isConfirmingDelete, handlers } = useThreadCardModeration(thread);

    const title = thread.title || thread.link?.title || "Untitled post";

    const hasReplied = thread.reply_status.replied > 0;
    const replyCount = thread.reply_status.replies;
    const replyCountLabel =
      replyCount === 1 ? `1 reply` : `${replyCount} replies`;

    const replyStatusLabel = hasReplied
      ? `${replyCountLabel} (including you!)`
      : replyCountLabel;

    const isInReview = thread.visibility === Visibility.review;
    const isAdmin = hasPermission(session, Permission.ADMINISTRATOR);

    return (
      <styled.div
        display="flex"
        flexDir="column"
        w="full"
        borderRadius="2xl"
        overflow="hidden"
        backgroundColor="white"
        style={{
          border: "1px solid #f1f2f4",
          boxShadow: "0 0.5px 2px rgba(0, 0, 0, 0.08)",
          transition: "all 0.2s ease-in-out",
          cursor: "pointer",
        }}
        onMouseEnter={(e) => {
          const el = e.currentTarget as HTMLElement;
          el.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.12)";
        }}
        onMouseLeave={(e) => {
          const el = e.currentTarget as HTMLElement;
          el.style.boxShadow = "0 1px 3px rgba(0, 0, 0, 0.08)";
        }}
      >
        <styled.div
          display="flex"
          alignItems="center"
          justifyContent="space-between"
          gap="3"
          px="4"
          pt="4"
          pb="2"
        >
          <styled.div display="flex" alignItems="center" gap="3" flex="1" minW="0">
            <ProfileHoverTooltip profile={thread.author}>
              <Link href={`/m/${thread.author.handle}`} style={{ textDecoration: "none" }}>
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
                    backgroundColor: isAdmin
                      ? TRAYA_COLORS.primary
                      : TRAYA_COLORS.tertiary,
                    color: isAdmin ? "white" : TRAYA_COLORS.primary,
                    border: "none",
                    cursor: "pointer",
                    transition: "opacity 0.2s ease-in-out",
                    flexShrink: 0,
                  }}
                >
                  {thread.author.handle.charAt(0).toUpperCase()}
                </styled.button>
              </Link>
            </ProfileHoverTooltip>

            <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
              <Link href={`/m/${thread.author.handle}`} style={{ textDecoration: "none" }}>
                <styled.p
                  fontSize="sm"
                  fontWeight="semibold"
                  color="fg.default"
                  style={{
                    margin: "0",
                    transition: "color 0.2s ease-in-out",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    cursor: "pointer",
                  }}
                >
                  @{thread.author.handle}
                </styled.p>
              </Link>
              <styled.p
                fontSize="xs"
                color="fg.muted"
                style={{
                  margin: "0",
                }}
              >
                {timestamp(new Date(thread.createdAt), false)}
              </styled.p>
            </styled.div>
          </styled.div>

          {session && (
            <styled.div
              style={{
                opacity: "0.7",
                transition: "opacity 0.2s ease-in-out",
                flexShrink: 0,
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = "1";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = "0.7";
              }}
            >
              <ThreadMenu thread={thread} />
            </styled.div>
          )}
        </styled.div>

        <styled.div
          display="flex"
          flexDir="column"
          gap="1"
          px="4"
          pt="2"
          pb="4"
        >
          <Link href={permalink} style={{ textDecoration: "none" }}>
            <styled.p
              fontSize="sm"
              color="fg.default"
              fontWeight="medium"
              style={{
                margin: "0",
                lineHeight: "1.6",
                whiteSpace: "pre-wrap",
                overflow: "hidden",
                display: "-webkit-box",
                WebkitLineClamp: 3,
                WebkitBoxOrient: "vertical",
                textOverflow: "ellipsis",
                cursor: "pointer",
              }}
            >
              {thread.description || title}
            </styled.p>
          </Link>

          {(!hideCategoryBadge && thread.category) || isInReview ? (
            <styled.div display="flex" gap="2.5" flexWrap="wrap">
              {!hideCategoryBadge && thread.category && (
                <CategoryBadge category={thread.category} />
              )}
              {isInReview && (
                <PostReviewBadge
                  isModerator={isModerator}
                  postId={thread.id}
                  onAccept={handlers.handleAcceptThread}
                  onEditAndAccept={handlers.handleEditAndAccept}
                  onDelete={handlers.handleConfirmDelete}
                  isConfirmingDelete={isConfirmingDelete}
                  onCancelDelete={handlers.handleCancelDelete}
                />
              )}
            </styled.div>
          ) : null}
        </styled.div>

        {!isInReview && session && (
          <styled.div
            display="flex"
            alignItems="center"
            gap="4"
            p="4"
            style={{
              borderTop: "1px solid var(--colors-border-default)/50",
              borderBottom: "1px solid var(--colors-border-default)/50",
            }}
          >
            <styled.div
              display="flex"
              alignItems="center"
              gap="2"
              style={{
                transition: "opacity 0.2s ease-in-out",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = "0.8";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = "1";
              }}
            >
              <LikeButton thread={thread} showCount />
            </styled.div>

            <Link href={permalink} title={replyStatusLabel}>
              <styled.div
                display="flex"
                alignItems="center"
                gap="1.5"
                color="fg.muted"
                style={{
                  transition: "color 0.2s ease-in-out",
                  fontSize: "14px",
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.color =
                    "var(--colors-fg-default)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.color =
                    "var(--colors-fg-muted)";
                }}
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

            <styled.div
              style={{
                marginLeft: "auto",
                opacity: "0.7",
                transition: "opacity 0.2s ease-in-out",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = "1";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = "0.7";
              }}
            >
              <CollectionMenu account={session} thread={thread} />
            </styled.div>
          </styled.div>
        )}
      </styled.div>
    );
  },
);

ThreadReferenceCard.displayName = "ThreadReferenceCard";
