"use client";

import { Controller, ControllerProps } from "react-hook-form";
import { match } from "ts-pattern";

import { Unready } from "src/components/site/Unready";

import { Permission, Thread, Visibility } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { hasPermission } from "@/utils/permissions";
import { CategoryBadge } from "@/components/category/CategoryBadge";
import { Byline } from "@/components/content/Byline";
import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { LinkCard } from "@/components/library/links/LinkCard";
import { CancelAction } from "@/components/site/Action/Cancel";
import { SaveAction } from "@/components/site/Action/Save";
import { PaginationControls } from "@/components/site/PaginationControls/PaginationControls";
import { TagBadgeList } from "@/components/tag/TagBadgeList";
import { Breadcrumbs } from "@/components/thread/Breadcrumbs";
import { PostReviewBadge } from "@/components/thread/PostReviewBadge";
import { ReplyBox } from "@/components/thread/ReplyBox/ReplyBox";
import { ReplyProvider } from "@/components/thread/ReplyContext";
import { ReplyList } from "@/components/thread/ReplyList/ReplyList";
import { ThreadDeletedAlert } from "@/components/thread/ThreadDeletedAlert";
import { ThreadMenu } from "@/components/thread/ThreadMenu/ThreadMenu";
import { TagListField } from "@/components/thread/ThreadTagList";
import { FormErrorText } from "@/components/ui/FormErrorText";
import { Heading } from "@/components/ui/heading";
import { HeadingInput } from "@/components/ui/heading-input";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import {
  DiscussionIcon,
  DiscussionParticipatingIcon,
} from "@/components/ui/icons/Discussion";
import { LikeIcon, LikeSavedIcon } from "@/components/ui/icons/Like";
import { VisibilityBadge } from "@/components/visibility/VisibilityBadge";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { HStack, LStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { useRouter } from "next/navigation";

import { Form, Props, useThreadScreen } from "./useThreadScreen";

export function ThreadScreen(props: Props) {
  const {
    ready,
    error,
    form,
    isEditing,
    isEmpty,
    resetKey,
    isModerator,
    isConfirmingDelete,
    data,
    handlers,
  } = useThreadScreen(props);

  if (!ready) {
    return <Unready error={error} />;
  }

  const { thread } = data;

  if (!thread) {
    return <Unready error={new Error("Thread data not found")} />;
  }

  const router = useRouter();
  const session = useSession(props.initialSession);
  const isAdmin = hasPermission(session, Permission.ADMINISTRATOR);

  return (
    <ReplyProvider>
      <LStack gap="0" width="full" height="screen" style={{ backgroundColor: "white" }}>
        {/* Mobile Header - Back Arrow + Post Title (Only on Mobile) */}
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
              (e.currentTarget as HTMLElement).style.backgroundColor = "rgba(0, 0, 0, 0.05)";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
            }}
          >
            <ArrowLeftIcon width="5" height="5" />
          </styled.button>
          <styled.h1 fontSize="md" fontWeight="semibold" color="fg.default">
            Post
          </styled.h1>
        </HStack>

        {/* Content Wrapper */}
        <styled.div
          flex="1"
          width="full"
          style={{
            overflowY: "auto",
            minHeight: 0,
          }}
        >
          {/* Desktop Breadcrumbs - Only on Desktop */}
          <styled.div display={{ base: "none", md: "block" }} px="4" py="4">
            <WStack alignItems="start">
              <Breadcrumbs
                thread={thread}
                channelID={props.channelID}
                channelName={props.channelName}
              />

              <HStack>
                {isEditing && (
                  <>
                    <CancelAction
                      type="button"
                      onClick={handlers.handleDiscardChanges}
                    >
                      Discard
                    </CancelAction>
                    <SaveAction type="submit" disabled={isEmpty}>
                      Save
                    </SaveAction>
                  </>
                )}
              </HStack>
            </WStack>
          </styled.div>

          <styled.form
            px="4"
            pt="4"
            pb="0"
            display="flex"
            flexDirection="column"
            gap="0"
            width="full"
            onSubmit={handlers.handleSave}
          >
            {/* Editing Controls - Mobile Only */}
            {isEditing && (
              <HStack gap="2" mb="2" className="md:hidden">
                <CancelAction
                  type="button"
                  onClick={handlers.handleDiscardChanges}
                >
                  Discard
                </CancelAction>
                <SaveAction type="submit" disabled={isEmpty}>
                  Save
                </SaveAction>
              </HStack>
            )}

            {thread.deletedAt !== undefined && (
              <ThreadDeletedAlert thread={thread} />
            )}

            {/* Post Header */}
            <HStack gap="3" alignItems="start" width="full" mb="1">
              <styled.button
                w="12"
                h="12"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                fontSize="lg"
                fontWeight="semibold"
                style={{
                  backgroundColor: isAdmin ? TRAYA_COLORS.primary : TRAYA_COLORS.secondary,
                  color: isAdmin ? "white" : TRAYA_COLORS.primary,
                  border: "none",
                  cursor: "pointer",
                  transition: "opacity 0.2s",
                  flexShrink: 0,
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.opacity = "0.8";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.opacity = "1";
                }}
              >
                {thread.author.handle.charAt(0).toUpperCase()}
              </styled.button>

              <VStack alignItems="start" gap="1" flex="1">
                <HStack gap="2" alignItems="center">
                  <styled.button
                    fontWeight="semibold"
                    color="fg.default"
                    style={{
                      background: "transparent",
                      border: "none",
                      cursor: "pointer",
                      padding: "0",
                      textDecoration: "none",
                      transition: "text-decoration 0.2s",
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.textDecoration = "underline";
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.textDecoration = "none";
                    }}
                  >
                    @{thread.author.handle}
                  </styled.button>
                  {isAdmin && (
                    <styled.span
                      px="2"
                      py="0.5"
                      fontSize="xs"
                      fontWeight="medium"
                      rounded="full"
                      style={{
                        backgroundColor: `${TRAYA_COLORS.primary}10`,
                        color: TRAYA_COLORS.primary,
                      }}
                    >
                      Admin
                    </styled.span>
                  )}
                </HStack>
                <styled.span fontSize="sm" color="fg.muted">
                  {new Date(thread.createdAt).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                  })}
                </styled.span>
              </VStack>

              {!isEditing && (
                <ThreadMenu thread={thread} editingEnabled movingEnabled onPinChange={handlers.handlePinChange} />
              )}
            </HStack>

            {/* Post Content */}
            <styled.div mb="1">
              {isEditing ? (
                <TitleInput name="title" control={form.control} />
              ) : (
                <styled.p
                  fontSize="sm"
                  color="fg.default"
                  style={{
                    whiteSpace: "pre-wrap",
                    lineHeight: "1.6",
                    margin: "0",
                    fontSize: "15px",
                  }}
                >
                  {thread.title}
                </styled.p>
              )}
            </styled.div>

            {/* Tags/Category */}
            <HStack gap="2" mb="4" flexWrap="wrap">
              {thread.category && (
                <styled.div
                  px="3"
                  py="1.5"
                  fontSize="sm"
                  fontWeight="medium"
                  rounded="full"
                  style={{
                    backgroundColor: TRAYA_COLORS.tertiary,
                    color: TRAYA_COLORS.primary,
                  }}
                >
                  {thread.category.name}
                </styled.div>
              )}
            </HStack>

            {/* Body Content */}
            {thread.link && <LinkCard link={thread.link} />}

            <ThreadBodyInput
              control={form.control}
              name="body"
              initialValue={thread.body}
              resetKey={resetKey}
              disabled={!isEditing}
              handleEmptyStateChange={handlers.handleEmptyStateChange}
            />

            {/* Actions */}
            <HStack
              gap="6"
              py="3"
              borderTopWidth="thin"
              borderTopColor="border.default"
              mt="4"
              style={{
                borderTopColor: "rgba(0, 0, 0, 0.05)",
              }}
            >
              <styled.button
                display="flex"
                alignItems="center"
                gap="2"
                fontSize="sm"
                fontWeight="medium"
                style={{
                  background: "transparent",
                  border: "none",
                  cursor: "pointer",
                  padding: "0",
                  color: thread.likes.liked ? TRAYA_COLORS.heart : "var(--colors-fg-muted)",
                  transition: "color 0.2s",
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.color = TRAYA_COLORS.heart;
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.color = thread.likes.liked ? TRAYA_COLORS.heart : "var(--colors-fg-muted)";
                }}
              >
                <span>
                  {thread.likes.liked ? (
                    <LikeSavedIcon width="5" />
                  ) : (
                    <LikeIcon width="5" />
                  )}
                </span>
                <span>{thread.likes.likes}</span>
              </styled.button>

              <HStack gap="2" fontSize="sm" fontWeight="medium" color="fg.muted">
                <span>
                  <DiscussionIcon width="5" />
                </span>
                <span>{thread.reply_status.replies}</span>
              </HStack>

              {match(thread.visibility)
                .with(Visibility.published, () => null)
                .with(Visibility.review, () => (
                  <PostReviewBadge
                    isModerator={isModerator}
                    postId={thread.id}
                    onAccept={handlers.handleAcceptThread}
                    onEditAndAccept={handlers.handleEditAndAccept}
                    onDelete={handlers.handleConfirmDelete}
                    isConfirmingDelete={isConfirmingDelete}
                    onCancelDelete={handlers.handleCancelDelete}
                  />
                ))
                .otherwise(() => (
                  <VisibilityBadge visibility={thread.visibility} />
                ))}
            </HStack>

            <FormErrorText>{form.formState.errors.root?.message}</FormErrorText>
          </styled.form>

          {/* Comments Section */}
          <styled.div py="4" px="2" style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
            <styled.h3 fontSize="md" fontWeight="semibold" color="fg.default">
              Comments ({thread.reply_status.replies})
            </styled.h3>

            {data.thread.replies.total_pages > 1 && (
              <PaginationControls
                path={
                  props.channelID
                    ? `/channels/${props.channelID}/threads/${thread.slug}`
                    : `/t/${thread.slug}`
                }
                currentPage={data.thread.replies.current_page ?? 1}
                totalPages={data.thread.replies.total_pages}
                pageSize={data.thread.replies.page_size}
              />
            )}

            <ReplyList
              initialSession={props.initialSession}
              thread={thread}
              currentPage={data.thread.replies.current_page}
            />

            {thread.reply_status.replies === 0 && (
              <styled.p textAlign="center" py="8" color="fg.muted">
                No comments yet. Be the first to comment!
              </styled.p>
            )}

            {data.thread.replies.total_pages > 1 && (
              <PaginationControls
                path={
                  props.channelID
                    ? `/channels/${props.channelID}/threads/${thread.slug}`
                    : `/t/${thread.slug}`
                }
                currentPage={data.thread.replies.current_page ?? 1}
                totalPages={data.thread.replies.total_pages}
                pageSize={data.thread.replies.page_size}
              />
            )}
          </styled.div>
        </styled.div>

        <ReplyBox initialSession={props.initialSession} thread={thread} />
      </LStack>
    </ReplyProvider>
  );
}

type TitleInputProps = Omit<ControllerProps<Form>, "render">;

export function TitleInput({ control }: TitleInputProps) {
  return (
    <Controller<Form>
      render={({ field: { onChange, ...field }, formState, fieldState }) => {
        return (
          <>
            <HeadingInput
              id="title-input"
              placeholder="Thread title..."
              onValueChange={onChange}
              defaultValue={formState.defaultValues?.["title"]}
              {...field}
            />

            <FormErrorText>{fieldState.error?.message}</FormErrorText>
          </>
        );
      }}
      control={control}
      name="title"
    />
  );
}

type ThreadBodyInputProps = Omit<ControllerProps<Form>, "render"> & {
  initialValue: string;
  resetKey: string;
  handleEmptyStateChange: (isEmpty: boolean) => void;
};

function ThreadBodyInput({
  control,
  name,
  initialValue,
  resetKey,
  disabled,
  handleEmptyStateChange,
}: ThreadBodyInputProps) {
  return (
    <Controller<Form>
      render={({ field: { onChange } }) => {
        function handleChange(value: string, isEmpty: boolean) {
          handleEmptyStateChange(isEmpty);
          onChange(value);
        }

        return (
          <ContentComposer
            initialValue={initialValue}
            onChange={handleChange}
            resetKey={resetKey}
            disabled={disabled}
          />
        );
      }}
      control={control}
      name={name}
    />
  );
}

function ThreadStats({ thread }: { thread: Thread }) {
  const likeCount = thread.likes.likes;
  const likeLabel = likeCount === 1 ? "like" : "likes";
  const replyCount = thread.reply_status.replies;
  const replyLabel = replyCount === 1 ? "reply" : "replies";

  return (
    <HStack gap="4" color="fg.muted">
      <styled.span
        display="flex"
        gap="1"
        alignItems="center"
        title={thread.likes.liked ? "You liked this thread" : undefined}
      >
        <span>
          {thread.likes.liked ? (
            <LikeSavedIcon width="4" />
          ) : (
            <LikeIcon width="4" />
          )}
        </span>
        <span>
          {likeCount} {likeLabel}
        </span>
      </styled.span>

      <styled.span
        display="flex"
        gap="1"
        alignItems="center"
        title={
          thread.reply_status.replied
            ? "You have replied to this thread"
            : undefined
        }
      >
        <span>
          {thread.reply_status.replied ? (
            <DiscussionParticipatingIcon width="4" />
          ) : (
            <DiscussionIcon width="4" />
          )}
        </span>
        <span>
          {replyCount} {replyLabel}
        </span>
      </styled.span>
    </HStack>
  );
}
