import Link from "next/link";
import { memo, useState } from "react";

import {
  Permission,
  ThreadReference,
  Visibility,
} from "src/api/openapi-schema";
import { useSession } from "src/auth";

import { styled } from "@/styled-system/jsx";
import { formatDateTime } from "@/utils/date";
import { hasPermission } from "@/utils/permissions";
import { getAvatarColor, getAuthorAvatarStyle } from "@/utils/avatar-colors";

import { CategoryBadge } from "../category/CategoryBadge";
import { PostReviewBadge } from "../thread/PostReviewBadge";
import { ThreadMenu } from "../thread/ThreadMenu/ThreadMenu";
import { DeleteConfirmSheet } from "../thread/ThreadMenu/DeleteConfirmSheet";
import {
  MessageSquareIcon,
  MessageSquareTextIcon,
} from "../ui/icons/MessageSquare";

import { PollCard } from "../poll/PollCard";
import { PostImages } from "./PostImages/PostImages";
import { ReplyChips } from "../thread/ReplyChips/ReplyChips";
import { usePostReplyChip } from "../thread/ReplyChips/usePostReplyChip";
import { ActionPill } from "./ActionPill";
import { BookmarkPill } from "./BookmarkPill";
import { ExpandableText } from "./ExpandableText";
import { FeedActionLike } from "./FeedActionLike";
import { useThreadCardModeration } from "./useThreadCardModeration";
import { ProfileHoverTooltip } from "./ProfileHoverTooltip";
import { SharedThreadCard } from "./SharedThreadCard";
// import { ShareThreadButton } from "./ShareThreadButton";

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

    const isSharedCard = thread.reference_post_id !== undefined;

    const { isConfirmingDelete, handlers } = useThreadCardModeration(thread);
    const [showDeleteSheet, setShowDeleteSheet] = useState(false);
    const chipCandidates = thread.quick_reply_chips?.candidates ?? [];
    const { postChip, isPosting, posted } = usePostReplyChip(thread);

    const title = thread.title || thread.link?.title || "Untitled post";

    const hasReplied = (thread.reply_status?.replied ?? 0) > 0;
    const replyCount = thread.reply_status?.replies ?? 0;
    const replyCountLabel =
      replyCount === 1 ? `1 reply` : `${replyCount} replies`;

    const replyStatusLabel = hasReplied
      ? `${replyCountLabel} (including you!)`
      : replyCountLabel;

    const isInReview = thread.visibility === Visibility.review;
    const isAdmin = hasPermission(session, Permission.ADMINISTRATOR);
    const showReviewState = isInReview && isAdmin;

    const bodyImages = [
      ...(thread.body ?? "").matchAll(/<img[^>]+src=["']([^"']+)["']/gi),
    ].map((m) => m[1] as string);

    const meta = thread.meta as Record<string, unknown> | undefined;
    const isPoll = meta?.["is_poll"] === true;
    const pollOptionDefs = isPoll
      ? (meta?.["poll_options"] as Array<{ id: string; text: string }> | undefined) ?? []
      : [];

    if (isSharedCard) {
      return <SharedThreadCard thread={thread} channelID={channelID} />;
    }

    return (
      <>
      <styled.div
        display="flex"
        flexDir="column"
        w="full"
        overflow="hidden"
        backgroundColor="white"
        style={{
          cursor: "pointer",
        }}
      >
        <styled.div
          display="flex"
          alignItems="center"
          justifyContent="space-between"
          gap="3"
          px="5"
          pt="7"
          pb="3"
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
                    ...getAuthorAvatarStyle(thread.author.handle, session?.handle),
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
                  fontWeight="bold"
                  style={{
                    margin: "0",
                    color: "#404040",
                    fontSize: "14px",
                    lineHeight: "20px",
                    transition: "color 0.2s ease-in-out",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                    cursor: "pointer",
                  }}
                >
                  {thread.author.name || `@${thread.author.handle}`}
                </styled.p>
              </Link>
              <styled.p
                fontSize="xs"
                color="fg.muted"
                style={{
                  margin: "0",
                }}
              >
                {formatDateTime(thread.createdAt)}
              </styled.p>
            </styled.div>
          </styled.div>

          {session && (
            <styled.div display="flex" alignItems="center" gap="3" flexShrink="0">
              <BookmarkPill account={session} thread={thread} />
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
                <ThreadMenu
                  thread={thread}
                  channelID={channelID}
                  onRequestDelete={() => setShowDeleteSheet(true)}
                />
              </styled.div>
            </styled.div>
          )}
        </styled.div>

        <styled.div
          display="flex"
          flexDir="column"
          gap="3"
          px="5"
          pt="0"
          pb="0"
        >
          {isPoll ? (
            <>
              <Link href={permalink} style={{ textDecoration: "none" }}>
                {thread.description && (
                  <styled.p
                    fontSize="sm"
                    fontWeight="semibold"
                    color="fg.default"
                    style={{ margin: "0 0 8px 0", lineHeight: "1.5" }}
                  >
                    {thread.description}
                  </styled.p>
                )}
              </Link>
              <PollCard threadMark={thread.slug} optionDefs={pollOptionDefs} />
            </>
          ) : (
            <>
              <ExpandableText
                text={thread.body || thread.description || title}
                fallback={thread.description || title}
                clampLines={3}
                href={permalink}
              />

              {bodyImages.length > 0 && (
                <styled.div mt="2" w="full">
                  {/* Only a lone image links through to the thread — wrapping a
                      carousel would swallow its swipes and indicator clicks. */}
                  {bodyImages.length === 1 ? (
                    <Link href={permalink} style={{ textDecoration: "none" }}>
                      <PostImages images={bodyImages} />
                    </Link>
                  ) : (
                    <PostImages images={bodyImages} />
                  )}
                </styled.div>
              )}
            </>
          )}

          {(!hideCategoryBadge && (thread.category || (thread.tags && thread.tags.length > 0))) || showReviewState ? (
            <styled.div display="flex" gap="1" flexWrap="wrap" alignItems="center">
              {!hideCategoryBadge && thread.category && (
                <CategoryBadge category={thread.category} />
              )}
              {!hideCategoryBadge && thread.tags?.map((tag) => (
                <styled.div
                  key={tag.name}
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  style={{
                    height: "19px",
                    padding: "0 8px",
                    borderRadius: "48px",
                    backgroundColor: "#ecf5f8",
                  }}
                >
                  <styled.span
                    fontWeight="bold"
                    style={{
                      color: "#404040",
                      fontSize: "10px",
                      lineHeight: "12px",
                      letterSpacing: "0.4px",
                      textTransform: "uppercase",
                    }}
                  >
                    {tag.name}
                  </styled.span>
                </styled.div>
              ))}
              {showReviewState && (
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

        {!showReviewState && session && (
          <styled.div
            display="flex"
            alignItems="center"
            gap="2"
            px="5"
            pt="5"
            pb="6"
          >
            <FeedActionLike thread={thread} />

            <Link href={permalink} title={replyStatusLabel} aria-label="Comments">
              <ActionPill
                as="span"
                ariaLabel="Comments"
                title={replyStatusLabel}
                count={replyCount}
                icon={
                  replyCount > 0 ? (
                    <MessageSquareTextIcon width="5" height="5" />
                  ) : (
                    <MessageSquareIcon width="5" height="5" />
                  )
                }
              />
            </Link>

            {/* Share hidden by request; component kept for when it returns. */}
            {/* <ShareThreadButton thread={thread} channelID={channelID} /> */}
          </styled.div>
        )}

        {!showReviewState && session && chipCandidates.length > 0 && (
          <styled.div px="4" pb="3">
            <ReplyChips
              candidates={chipCandidates}
              onPick={postChip}
              isPosting={isPosting}
              posted={posted}
            />
          </styled.div>
        )}
      </styled.div>

      {session && (
        <DeleteConfirmSheet
          isOpen={showDeleteSheet}
          onOpenChange={setShowDeleteSheet}
          onConfirm={handlers.handleDelete}
        />
      )}
      </>
    );
  },
);

ThreadReferenceCard.displayName = "ThreadReferenceCard";
