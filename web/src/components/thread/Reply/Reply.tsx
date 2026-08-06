import Link from "next/link";
import { Controller, ControllerProps } from "react-hook-form";

import { Reply as ReplyType, Thread } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { CohortBadge } from "@/components/category/CohortBadge";
import { MemberBadge } from "@/components/member/MemberBadge/MemberBadge";
import { CancelAction } from "@/components/site/Action/Cancel";
import { SaveAction } from "@/components/site/Action/Save";
import { Timestamp } from "@/components/site/Timestamp";
import { ReplyIcon } from "@/components/ui/icons/Reply";
import { HStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { hstack } from "@/styled-system/patterns";
import { getAvatarColor, getAuthorAvatarStyle } from "@/utils/avatar-colors";

import { PostReviewBadge } from "../PostReviewBadge";
import { ReactList } from "../ReactList/ReactList";
import { ReplyMenu } from "../ReplyMenu/ReplyMenu";

import { ReplyToButton } from "./ReplyToButton";
import { useFragmentScroll } from "./useFragmentScroll";
import { Form, Props, useReply } from "./useReply";

export function Reply(props: Props) {
  const session = useSession();
  const {
    isEmpty,
    isEditing,
    isEditingInReview,
    canManageReplies,
    resetKey,
    form,
    isConfirmingDelete,
    handlers,
  } = useReply(props);
  const isTargeted = useFragmentScroll(props.reply.id);

  const { initialSession, thread, reply, currentPage } = props;

  const isInReview = reply.visibility === "review";

  return (
    <styled.div
      id={reply.id}
      width="full"
      data-targeted={isTargeted || undefined}
      _target={{
        scrollMarginTop: { base: "0", md: "20" },
        animation: "target-pulse",
      }}
      css={{
        "&[data-targeted]": { animation: "target-pulse" },
      }}
      style={{
        backgroundColor: isInReview ? "#fff7ed" : "#f7f7f7",
        borderRadius: "12px",
        padding: "12px",
      }}
    >
      <styled.form
        display="flex"
        flexDirection="column"
        gap="2"
        onSubmit={handlers.handleSave}
      >
        {/* Comment header */}
        <HStack gap="1.5" alignItems="start" width="full">
          <Link href={`/m/${reply.author.handle}`} style={{ textDecoration: "none" }}>
            <styled.div
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              rounded="full"
              fontWeight="semibold"
              style={{
                width: "22px",
                height: "22px",
                fontSize: "10.4px",
                ...getAuthorAvatarStyle(reply.author.handle, session?.handle),
              }}
            >
              {(reply.author.name?.charAt(0) || reply.author.handle.charAt(0)).toUpperCase()}
            </styled.div>
          </Link>

          <VStack alignItems="start" gap="0.5" flex="1" minW="0">
            <HStack gap="2" alignItems="center">
              <Link href={`/m/${reply.author.handle}`} style={{ textDecoration: "none" }}>
                <styled.span
                  fontWeight="bold"
                  style={{ fontSize: "14px", lineHeight: "20px", color: "#404040" }}
                >
                  {reply.author.name || `@${reply.author.handle}`}
                </styled.span>
              </Link>
              {reply.cohort_channel && <CohortBadge channel={reply.cohort_channel} />}
            </HStack>
            <styled.span style={{ fontSize: "10px", lineHeight: "12px", color: "#5c5c5c", letterSpacing: "0.4px" }}>
              <Timestamp created={reply.createdAt} />
            </styled.span>
          </VStack>

          {isEditing && (
            <HStack>
              <CancelAction type="button" onClick={handlers.handleDiscardChanges}>
                Discard
              </CancelAction>
              <SaveAction type="submit" disabled={isEmpty}>
                {isEditingInReview ? "Accept" : "Save"}
              </SaveAction>
            </HStack>
          )}
        </HStack>

        {reply.reply_to && <InReplyTo to={reply.reply_to} thread={thread} />}

        <ReplyBodyInput
          control={form.control}
          name="body"
          initialValue={reply.body}
          resetKey={resetKey}
          disabled={!isEditing}
          handleEmptyStateChange={handlers.handleEmptyStateChange}
        />
      </styled.form>

      {!isEditing && (
        <WStack alignItems="center" mt="2">
          <ReplyToButton thread={thread} reply={reply} />

          <HStack gap="1" alignItems="center">
            <ReactList
              initialSession={initialSession}
              thread={thread}
              reply={reply}
              currentPage={currentPage}
            />
            <ReplyMenu
              thread={thread}
              reply={reply}
              currentPage={currentPage}
              onEdit={handlers.handleSetEditing}
            />
          </HStack>
        </WStack>
      )}

      {isInReview && (
        <PostReviewBadge
          isModerator={canManageReplies}
          postId={reply.id}
          onAccept={handlers.handleAcceptReply}
          onEditAndAccept={handlers.handleSetEditingInReview}
          onDelete={handlers.handleConfirmDelete}
          isConfirmingDelete={isConfirmingDelete}
          onCancelDelete={handlers.handleCancelDelete}
        />
      )}
    </styled.div>
  );
}

type ReplyBodyInputProps = Omit<ControllerProps<Form>, "render"> & {
  initialValue: string;
  resetKey: string;
  handleEmptyStateChange: (isEmpty: boolean) => void;
};

function ReplyBodyInput({
  control,
  name,
  initialValue,
  resetKey,
  disabled,
  handleEmptyStateChange,
}: ReplyBodyInputProps) {
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

function InReplyTo({ to, thread }: { to: ReplyType; thread: Thread }) {
  // figure out if the reply-to is on the current page, then  do a fragment link
  // if on same page, otherwise use /t/locate to navigate to the right page.
  const isOnCurrentPage = thread.replies.replies.some((r) => r.id === to.id);
  const permalink = isOnCurrentPage ? `#${to.id}` : `/t/locate/${to.id}`;

  // NOTE: because nextjs does some weird shit, we gotta use a normal anchor
  // for fragment navigation, otherwise it breaks :target etc for some reason.
  const AnchorComponent = isOnCurrentPage ? styled.a : Link;

  return (
    <WStack
      gap="1"
      fontSize="xs"
      color="fg.muted"
      px="2"
      py="1"
      borderRadius="md"
      bgColor="bg.subtle"
      w="full"
      minW="0"
    >
      <AnchorComponent
        href={permalink}
        className={hstack({
          minW: "0",
          flexShrink: "1",
        })}
      >
        <ReplyIcon w="4" minW="4" />
        <styled.span
          minW="0"
          overflow="hidden"
          textOverflow="ellipsis"
          whiteSpace="nowrap"
          lineClamp="1"
        >
          “{to.description}”
        </styled.span>
      </AnchorComponent>

      <HStack flexShrink="0" minW="0">
        <MemberBadge
          profile={to.author}
          size="xs"
          name="handle"
          avatar="visible"
        />
        <AnchorComponent href={permalink}>
          <Timestamp created={to.createdAt} />
        </AnchorComponent>
      </HStack>
    </WStack>
  );
}
