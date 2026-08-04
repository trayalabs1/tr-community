"use client";

import Link from "next/link";
import { Controller, ControllerProps } from "react-hook-form";
import { match } from "ts-pattern";

import { Unready } from "src/components/site/Unready";

import { Permission, ThreadReference, Visibility } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { LinkCard } from "@/components/library/links/LinkCard";
import { PollCard } from "@/components/poll/PollCard";
import { LikeButton } from "@/components/post/LikeButton/LikeButton";
import { ExpandableText } from "@/components/post/ExpandableText";
import { BookmarkPill } from "@/components/post/BookmarkPill";
import { CategoryBadge } from "@/components/category/CategoryBadge";
import {
  MessageSquareIcon,
  MessageSquareTextIcon,
} from "@/components/ui/icons/MessageSquare";
import { CancelAction } from "@/components/site/Action/Cancel";
import { SaveAction } from "@/components/site/Action/Save";
import { HeaderWithBackArrow } from "@/components/site/Header";
import { PaginationControls } from "@/components/site/PaginationControls/PaginationControls";
import { Breadcrumbs } from "@/components/thread/Breadcrumbs";
import { PostReviewBadge } from "@/components/thread/PostReviewBadge";
import { ReplyBox } from "@/components/thread/ReplyBox/ReplyBox";
import { ReplyProvider } from "@/components/thread/ReplyContext";
import { ReplyList } from "@/components/thread/ReplyList/ReplyList";
import { ThreadDeletedAlert } from "@/components/thread/ThreadDeletedAlert";
import { ThreadMenu } from "@/components/thread/ThreadMenu/ThreadMenu";
import { TagListField } from "@/components/thread/ThreadTagList";
import { FormErrorText } from "@/components/ui/FormErrorText";
import { HeadingInput } from "@/components/ui/heading-input";
import { VisibilityBadge } from "@/components/visibility/VisibilityBadge";
import { HStack, LStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { getAvatarColor, getAuthorAvatarStyle } from "@/utils/avatar-colors";
import { hasPermission } from "@/utils/permissions";

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

  const session = useSession(props.initialSession);
  const isAdmin = hasPermission(session, Permission.ADMINISTRATOR);

  const bodyImages = [
    ...(thread.body ?? "").matchAll(/<img[^>]+src=["']([^"']+)["']/gi),
  ].map((m) => m[1] as string);

  const threadMeta = thread.meta as Record<string, unknown> | undefined;
  const isPoll = threadMeta?.["is_poll"] === true;
  const pollOptionDefs = isPoll
    ? (threadMeta?.["poll_options"] as Array<{ id: string; text: string }> | undefined) ?? []
    : [];

  return (
    <ReplyProvider>
      <LStack gap="0" width="full" height="screen" maxW="[600px]" mx="auto" bg="bg.surfaceWhite">
        <HeaderWithBackArrow
          title=""
          mobileOnly
          isSticky
        />

        {/* Content Wrapper */}
        <styled.div
          flex="1"
          width="full"
          pb={{ base: "28", md: "0" }}
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
            px="5"
            pt="7"
            pb="7"
            display="flex"
            flexDirection="column"
            gap="0"
            width="full"
            bg="white"
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
            <HStack gap="3" alignItems="center" width="full" mb="3">
              <Link href={`/m/${thread.author.handle}`} style={{ textDecoration: "none" }}>
                <styled.div
                  rounded="full"
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  fontWeight="semibold"
                  style={{
                    width: "38px",
                    height: "38px",
                    fontSize: "18px",
                    ...getAuthorAvatarStyle(thread.author.handle, session?.handle),
                    cursor: "pointer",
                    flexShrink: 0,
                  }}
                >
                  {(thread.author.name?.charAt(0) || thread.author.handle.charAt(0)).toUpperCase()}
                </styled.div>
              </Link>

              <VStack alignItems="start" gap="0.5" flex="1" minW="0">
                <Link href={`/m/${thread.author.handle}`} style={{ textDecoration: "none" }}>
                  <styled.span
                    fontWeight="bold"
                    style={{
                      fontSize: "14px",
                      lineHeight: "20px",
                      color: TRAYA_COLORS.primary,
                    }}
                  >
                    {thread.author.name || `@${thread.author.handle}`}
                  </styled.span>
                </Link>
                <styled.span style={{ fontSize: "12px", lineHeight: "16px", color: "#999999", letterSpacing: "0.24px" }}>
                  {new Date(thread.createdAt).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                  })}
                </styled.span>
              </VStack>

              {!isEditing && session && (
                <HStack gap="3" alignItems="center" flexShrink="0">
                  <BookmarkPill account={session} thread={thread as unknown as ThreadReference} />
                  <ThreadMenu thread={thread} editingEnabled movingEnabled onPinChange={handlers.handlePinChange} />
                </HStack>
              )}
            </HStack>

            {/* Post Content */}
            {!isPoll && (
              <styled.div mb="3" width="full">
                {isEditing ? (
                  <TitleInput name="title" control={form.control} />
                ) : (
                  <ExpandableText
                    text={thread.body || thread.description || thread.title}
                    fallback={thread.description || thread.title}
                    clampLines={6}
                  />
                )}
              </styled.div>
            )}

            {/* Images from the post body. ExpandableText renders text only, so
                they are pulled out and shown here. */}
            {!isPoll && !isEditing && bodyImages.length > 0 && (
              <styled.div
                display="flex"
                flexDirection="column"
                gap="2"
                mb="4"
                w="full"
              >
                {bodyImages.map((src) => (
                  <styled.img
                    key={src}
                    src={src}
                    alt=""
                    w="full"
                    h="auto"
                    borderRadius="lg"
                    style={{ objectFit: "cover" }}
                  />
                ))}
              </styled.div>
            )}

            {/* Tags/Category */}
            {(thread.category || (thread.tags && thread.tags.length > 0)) && (
              <HStack gap="1" mb="4" flexWrap="wrap" alignItems="center">
                {thread.category && <CategoryBadge category={thread.category} />}
                {thread.tags?.map((tag) => (
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
                      style={{ color: "#404040", fontSize: "10px", lineHeight: "12px", letterSpacing: "0.4px", textTransform: "uppercase" }}
                    >
                      {tag.name}
                    </styled.span>
                  </styled.div>
                ))}
              </HStack>
            )}

            {/* Body Content */}
            {thread.link && <LinkCard link={thread.link} />}

            {isPoll ? (
              <styled.div mb="4">
                {thread.description && (
                  <styled.p
                    fontSize="md"
                    fontWeight="semibold"
                    color="fg.default"
                    mb="3"
                    style={{ lineHeight: "1.5" }}
                  >
                    {thread.description}
                  </styled.p>
                )}
                <PollCard threadMark={thread.slug} optionDefs={pollOptionDefs} />
              </styled.div>
            ) : (
              // Read mode shows the body via ExpandableText above; the rich
              // ContentComposer only mounts when editing.
              isEditing && (
                <ThreadBodyInput
                  control={form.control}
                  name="body"
                  initialValue={thread.body}
                  resetKey={resetKey}
                  disabled={!isEditing}
                  handleEmptyStateChange={handlers.handleEmptyStateChange}
                />
              )
            )}

            {/* Actions */}
            <HStack gap="2.5" mt="4" alignItems="center">
              <LikeButton thread={thread} showCount />

              <HStack gap="1" alignItems="center">
                <styled.div display="flex" alignItems="center" justifyContent="center" w="7" h="7" color="fg.muted">
                  {thread.reply_status.replies > 0 ? (
                    <MessageSquareTextIcon width="5" height="5" />
                  ) : (
                    <MessageSquareIcon width="5" height="5" />
                  )}
                </styled.div>
                <styled.span style={{ fontSize: "12px", fontWeight: 500, color: "#404040", letterSpacing: "0.24px" }}>
                  {thread.reply_status.replies}
                </styled.span>
              </HStack>

              {match(thread.visibility)
                .with(Visibility.published, () => null)
                .with(Visibility.review, () =>
                  isAdmin ? (
                    <PostReviewBadge
                      isModerator={isModerator}
                      postId={thread.id}
                      onAccept={handlers.handleAcceptThread}
                      onEditAndAccept={handlers.handleEditAndAccept}
                      onDelete={handlers.handleConfirmDelete}
                      isConfirmingDelete={isConfirmingDelete}
                      onCancelDelete={handlers.handleCancelDelete}
                    />
                  ) : null,
                )
                .otherwise(() => (
                  <VisibilityBadge visibility={thread.visibility} />
                ))}
            </HStack>

            <FormErrorText>{form.formState.errors.root?.message}</FormErrorText>
          </styled.form>

          {/* Comments Section */}
          <styled.div py="4" px="4" bg="bg.surfaceWhite" style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
            <styled.h3 style={{ fontSize: "18px", lineHeight: "22px", fontWeight: 600, color: "#404040" }}>
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

        <styled.div
          flexShrink="0"
          width="full"
          position={{ base: "fixed", md: "sticky" }}
          bottom="0"
          left="0"
          right="0"
          zIndex="sticky"
          maxW={{ base: "full", md: "[600px]" }}
          mx="auto"
        >
          <ReplyBox initialSession={props.initialSession} thread={thread} />
        </styled.div>
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

