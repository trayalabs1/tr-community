import Link from "next/link";
import { Controller, ControllerProps } from "react-hook-form";

import { Anchor } from "src/components/site/Anchor";

import { Send } from "lucide-react";

import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { MemberIdent } from "@/components/member/MemberBadge/MemberIdent";
import { Admonition } from "@/components/ui/admonition";
import { IconButton } from "@/components/ui/icon-button";
import { CloseIcon } from "@/components/ui/icons/Close";
import { DiscussionIcon } from "@/components/ui/icons/Discussion";
import { css } from "@/styled-system/css";
import { HStack, LStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { timestamp } from "@/utils/date";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { ReplyChips } from "../ReplyChips/ReplyChips";
import { usePostReplyChip } from "../ReplyChips/usePostReplyChip";
import { useReplyContext } from "../ReplyContext";

import { Form, Props, useReplyBox } from "./useReplyBox";

export function ReplyBox(props: Props) {
  const { replyTo, clearReplyTo } = useReplyContext();
  const {
    isLoggedIn,
    isEmpty,
    isLoading,
    form,
    resetKey,
    postedReply,
    handlers,
  } = useReplyBox(props);
  const chipCandidates = props.thread.quick_reply_chips?.candidates ?? [];
  const { postChip, isPosting: isPostingChip, posted: chipPosted } = usePostReplyChip(props.thread);

  if (!isLoggedIn) {
    return <LoginToReply />;
  }

  return (
    <VStack w="full" gap="2" alignItems="stretch">
      <Admonition
        value={!!postedReply}
        onChange={handlers.handleReplyPostedAdmonitionClose}
      >
        {postedReply && (
          <LStack h="full" justifyContent="center">
            <styled.p fontSize="sm" color="fg.muted">
              Your reply has been posted on{" "}
              <Link
                className={css({
                  color: "fg.emphasized",
                  _hover: { textDecoration: "underline" },
                })}
                href={postedReply.permalink}
                onClick={handlers.handleReplyNavigation}
              >
                page {postedReply.pageNumber}
              </Link>
              .
            </styled.p>
          </LStack>
        )}
      </Admonition>

      <styled.form
        display="flex"
        flexDirection="column"
        gap="2"
        width="full"
        px="4"
        py="3"
        bg="white"
        style={{ borderTop: `1px solid ${TRAYA_COLORS.neutral.border}` }}
        onSubmit={handlers.handleSubmit}
      >
        {replyTo && (
          <WStack py="1" px="2" borderRadius="md" bgColor="bg.muted">
            <HStack gap="1" fontSize="sm" color="fg.muted">
              <styled.span>Replying&nbsp;to</styled.span>
              <MemberIdent
                profile={replyTo.reply.author}
                name="handle"
                size="xs"
              />
              <styled.a href={`#${replyTo.reply.id}`}>
                {timestamp(replyTo.reply.createdAt)}
              </styled.a>
            </HStack>

            <IconButton
              type="button"
              size="xs"
              variant="ghost"
              aria-label="Clear reply-to"
              onClick={clearReplyTo}
            >
              <CloseIcon />
            </IconButton>
          </WStack>
        )}

        {chipCandidates.length > 0 && (
          <ReplyChips
            candidates={chipCandidates}
            onPick={postChip}
            isPosting={isPostingChip}
            posted={chipPosted}
          />
        )}

        <HStack gap="3" alignItems="center" width="full">
          <styled.div
            flex="1"
            minW="0"
            display="flex"
            alignItems="center"
            style={{
              border: `1px solid ${TRAYA_COLORS.neutral.border}`,
              borderRadius: "16px",
              padding: "2px 12px",
              minHeight: "40px",
            }}
          >
            <ReplyBodyInput
              name="body"
              control={form.control}
              handleEmptyStateChange={handlers.handleEmptyStateChange}
              resetKey={resetKey}
            />
          </styled.div>

          <styled.button
            type="submit"
            disabled={isLoading || isEmpty}
            display="flex"
            alignItems="center"
            justifyContent="center"
            flexShrink="0"
            aria-label="Send reply"
            style={{
              width: "40px",
              height: "40px",
              borderRadius: "10px",
              border: "none",
              cursor: isLoading || isEmpty ? "default" : "pointer",
              backgroundColor: isEmpty ? "#c9c9c9" : TRAYA_COLORS.primary,
              transition: "background-color 0.2s",
            }}
          >
            <Send size={20} color="white" />
          </styled.button>
        </HStack>
      </styled.form>
    </VStack>
  );
}

type ReplyBodyInputProps = Omit<ControllerProps<Form>, "render"> & {
  handleEmptyStateChange: (isEmpty: boolean) => void;
  resetKey: string;
};

function ReplyBodyInput({
  control,
  name,
  handleEmptyStateChange,
  resetKey,
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
            onChange={handleChange}
            resetKey={resetKey}
            placeholder="Write your heart out....."
          />
        );
      }}
      control={control}
      name={name}
    />
  );
}

function LoginToReply() {
  return (
    <HStack
      w="full"
      p="8"
      borderRadius="xl"
      bgColor="border.muted"
      justifyContent="center"
    >
      <DiscussionIcon width="4" />

      <p>
        Please <Anchor href="/register">sign up or log in</Anchor> to reply
      </p>
    </HStack>
  );
}
