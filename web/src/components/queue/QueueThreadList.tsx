import { ThreadReference } from "@/api/openapi-schema";
import { MemberBadge } from "@/components/member/MemberBadge/MemberBadge";
import { PostReviewBadge } from "@/components/thread/PostReviewBadge";
import { ThreadMenu } from "@/components/thread/ThreadMenu/ThreadMenu";
import { CategoryBadge } from "@/components/category/CategoryBadge";
import { Timestamp } from "@/components/site/Timestamp";
import { Card, CardRows } from "@/components/ui/rich-card";
import { HStack, WStack, VStack, styled } from "@/styled-system/jsx";
import { useSession } from "@/auth";
import { hasPermission } from "@/utils/permissions";
import { Permission } from "@/api/openapi-schema";
import { useThreadCardModeration } from "@/components/post/useThreadCardModeration";

export function QueueThreadList({
  threads,
}: {
  threads: ThreadReference[];
}) {
  if (threads.length === 0) {
    return <p>No threads waiting for review.</p>;
  }

  return (
    <CardRows>
      {threads.map((thread) => (
        <QueueThreadListItem key={thread.id} thread={thread} />
      ))}
    </CardRows>
  );
}

export function QueueThreadListItem({
  thread,
}: {
  thread: ThreadReference;
}) {
  const session = useSession();
  const isModerator = hasPermission(
    session,
    Permission.MANAGE_POSTS,
    Permission.ADMINISTRATOR,
  );

  const { isConfirmingDelete, handlers } = useThreadCardModeration(thread);

  const url = `/t/${thread.slug}`;
  const title = thread.title || thread.link?.title || "Untitled post";

  return (
    <Card
      key={thread.id}
      id={thread.id}
      shape="responsive"
      url={url}
      title={title}
      text={thread.description}
      controls={
        <WStack>
          <VStack gap="2" w="full">
            <HStack gap="2">
              <MemberBadge profile={thread.author} size="sm" />

              {thread.category && (
                <CategoryBadge category={thread.category} />
              )}

              <PostReviewBadge
                isModerator={isModerator}
                postId={thread.id}
                onAccept={handlers.handleAcceptThread}
                onEditAndAccept={handlers.handleEditAndAccept}
                onDelete={handlers.handleConfirmDelete}
                isConfirmingDelete={isConfirmingDelete}
                onCancelDelete={handlers.handleCancelDelete}
              />

              <Timestamp created={thread.createdAt} large />
            </HStack>

            {/* Channel name - if available */}
            {/* <styled.div fontSize="xs" color="fg.muted">
              Channel: <styled.span fontWeight="medium" color="fg.default">
                {(thread as any).channel_name || "Unknown Channel"}
              </styled.span>
            </styled.div> */}
          </VStack>

          <ThreadMenu thread={thread} editingEnabled movingEnabled />
        </WStack>
      }
    />
  );
}
