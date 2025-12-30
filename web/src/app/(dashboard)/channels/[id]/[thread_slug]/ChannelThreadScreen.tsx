"use client";

import Link from "next/link";

import { Unready } from "src/components/site/Unready";

import { useChannelThreadGet } from "@/api/openapi-client/channels";
import { Account, Channel, ThreadGetResponse } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { ReplyProvider } from "@/components/thread/ReplyContext";
import { ReplyList } from "@/components/thread/ReplyList/ReplyList";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";

type Props = {
  session?: Account;
  channel: Channel;
  channelId: string;
  slug: string;
  thread: ThreadGetResponse;
  initialPage: number;
};

export function ChannelThreadScreen(props: Props) {
  const session = useSession(props.session);

  const { data, error } = useChannelThreadGet(
    props.channelId,
    props.slug,
    {
      page: props.initialPage.toString(),
    },
    {
      swr: {
        fallbackData: props.thread,
        revalidateOnFocus: false,
      },
    }
  );

  if (error) {
    return <Unready error={error} />;
  }

  if (!data) {
    return <Unready />;
  }

  return (
    <ReplyProvider>
      <LStack gap="4" p="4">
        <VStack alignItems="start" gap="2" width="full">
          <HStack gap="2" fontSize="sm" color="fg.muted">
            <Link href={`/channels/${props.channel.id}`}>
              ← Back to {props.channel.name}
            </Link>
          </HStack>

          <Heading as="h1" size="2xl">
            {data.title}
          </Heading>

          <styled.div color="fg.muted" fontSize="sm">
            by {data.author.name}
          </styled.div>

          {data.body && (
            <styled.div
              width="full"
              dangerouslySetInnerHTML={{ __html: data.body }}
            />
          )}
        </VStack>

        <VStack alignItems="start" gap="4" width="full">
          <Heading as="h2" size="lg">
            Replies ({data.replies?.results || 0})
          </Heading>

          <ReplyList
            initialSession={props.session}
            thread={data}
            currentPage={data.replies?.current_page}
          />
        </VStack>
      </LStack>
    </ReplyProvider>
  );
}
