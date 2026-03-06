import Link from "next/link";

import { ThreadReference } from "@/api/openapi-schema";
import { MemberBadge } from "@/components/member/MemberBadge/MemberBadge";
import { Timestamp } from "@/components/site/Timestamp";
import { Card, CardRows } from "@/components/ui/rich-card";
import { HStack, WStack, VStack } from "@/styled-system/jsx";

export function PendingReplyThreadList({ threads }: { threads: ThreadReference[] }) {
  if (threads.length === 0) {
    return <p>No threads pending reply.</p>;
  }

  return (
    <CardRows>
      {threads.map((thread) => (
        <PendingReplyThreadListItem key={thread.id} thread={thread} />
      ))}
    </CardRows>
  );
}

function PendingReplyThreadListItem({ thread }: { thread: ThreadReference }) {
  const url = thread.channel_id
    ? `/channels/${thread.channel_id}/threads/${thread.slug}`
    : `/t/${thread.slug}`;
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
              <Timestamp created={thread.createdAt} large />
            </HStack>
          </VStack>
          <Link href={url}>
            <HStack
              gap="1"
              px="3"
              py="1.5"
              rounded="md"
              fontSize="sm"
              fontWeight="semibold"
              cursor="pointer"
              style={{
                backgroundColor: "var(--colors-accent-default)",
                color: "var(--colors-accent-fg)",
              }}
            >
              Reply
            </HStack>
          </Link>
        </WStack>
      }
    />
  );
}
