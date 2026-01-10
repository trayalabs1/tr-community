import Link from "next/link";
import { memo } from "react";

import {
  Permission,
  ThreadReference,
  Visibility,
} from "src/api/openapi-schema";
import { useSession } from "src/auth";
import { CollectionMenu } from "src/components/content/CollectionMenu/CollectionMenu";
import { MemberAvatar } from "src/components/member/MemberBadge/MemberAvatar";

import { styled } from "@/styled-system/jsx";
import { timestamp } from "@/utils/date";
import { hasPermission } from "@/utils/permissions";

import { CategoryBadge } from "../category/CategoryBadge";
import { PostReviewBadge } from "../thread/PostReviewBadge";
import { ThreadMenu } from "../thread/ThreadMenu/ThreadMenu";
import {
  DiscussionIcon,
  DiscussionParticipatingIcon,
} from "../ui/icons/Discussion";

import { LikeButton } from "./LikeButton/LikeButton";
import { useThreadCardModeration } from "./useThreadCardModeration";

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

    return (
      <styled.div
        display="flex"
        flexDir="column"
        w="full"
        // maxW="2xl"
        borderRadius="lg"
        overflow="hidden"
        backgroundColor="bg.default"
        style={{
          border: "1px solid var(--colors-border-subtle)",
          boxShadow: "0 1px 3px rgba(0, 0, 0, 0.08)",
          transition: "all 0.2s ease-in-out",
        }}
        onMouseEnter={(e) => {
          const el = e.currentTarget as HTMLElement;
          el.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.12)";
          el.style.borderColor = "var(--colors-border-default)";
        }}
        onMouseLeave={(e) => {
          const el = e.currentTarget as HTMLElement;
          el.style.boxShadow = "0 1px 3px rgba(0, 0, 0, 0.08)";
          el.style.borderColor = "var(--colors-border-subtle)";
        }}
      >
        <styled.div
          display="flex"
          alignItems="flex-start"
          justifyContent="space-between"
          gap="4"
          p="5"
          style={{
            borderBottom: "1px solid var(--colors-border-subtle)",
          }}
        >
          <Link href={`/m/${thread.author.handle}`}>
            <styled.div display="flex" alignItems="center" gap="3" flex="1">
              <styled.div
                style={{
                  transition: "transform 0.2s ease-in-out",
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.transform =
                    "scale(1.05)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.transform =
                    "scale(1)";
                }}
              >
                <MemberAvatar profile={thread.author} size="md" />
              </styled.div>
              <styled.div display="flex" flexDir="column" gap="1" flex="1">
                <styled.p
                  fontSize="sm"
                  fontWeight="semibold"
                  color="fg.default"
                  style={{
                    transition: "color 0.2s ease-in-out",
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.color =
                      "var(--colors-fg-subtle)";
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.color =
                      "var(--colors-fg-default)";
                  }}
                >
                  {thread.author.handle}
                </styled.p>
                <styled.p fontSize="xs" color="fg.muted">
                  {timestamp(new Date(thread.createdAt), false)}
                </styled.p>
              </styled.div>
            </styled.div>
          </Link>
          {session && (
            <styled.div
              style={{
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
              <ThreadMenu thread={thread} />
            </styled.div>
          )}
        </styled.div>

        <styled.div
          display="flex"
          flexDir="column"
          gap="4"
          p="5"
          flex="1"
          style={{
            borderBottom: "1px solid var(--colors-border-subtle)",
          }}
        >
          <Link href={permalink}>
            <styled.h3
              fontSize="lg"
              fontWeight="bold"
              color="fg.default"
              style={{
                lineHeight: "1.6",
                transition: "color 0.2s ease-in-out",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.color =
                  "var(--colors-fg-subtle)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.color =
                  "var(--colors-fg-default)";
              }}
            >
              {title}
            </styled.h3>
          </Link>

          {thread.description && (
            <styled.p
              fontSize="sm"
              color="fg.muted"
              style={{
                lineHeight: "1.6",
                display: "-webkit-box",
                WebkitLineClamp: 2,
                WebkitBoxOrient: "vertical",
                overflow: "hidden",
              }}
            >
              {thread.description}
            </styled.p>
          )}

          {!hideCategoryBadge && thread.category && (
            <styled.div display="flex" gap="2" flexWrap="wrap">
              <CategoryBadge category={thread.category} />
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
          )}
        </styled.div>

        {!isInReview && session && (
          <styled.div
            display="flex"
            alignItems="center"
            justifyContent="space-between"
            gap="4"
            p="5"
            backgroundColor="bg.subtle"
          >
            <styled.div display="flex" alignItems="center" gap="8">
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
                  gap="2"
                  color="fg.muted"
                  style={{
                    transition: "color 0.2s ease-in-out",
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
            </styled.div>

            <styled.div
              style={{
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
