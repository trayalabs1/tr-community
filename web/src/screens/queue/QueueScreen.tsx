"use client";

import Link from "next/link";
import { useChannelList } from "@/api/openapi-client/channels";
import { useNodeList } from "@/api/openapi-client/nodes";
import { useThreadList } from "@/api/openapi-client/threads";
import { Visibility } from "@/api/openapi-schema";
import { QueueNodeList } from "@/components/queue/QueueNodeList";
import { QueueThreadList } from "@/components/queue/QueueThreadList";
import { Unready } from "@/components/site/Unready";
import { Heading } from "@/components/ui/heading";
import { VStack, LStack, HStack, styled } from "@/styled-system/jsx";
import { getAssetURL } from "@/utils/asset";
import { MembersIcon } from "@/components/ui/icons/Members";

export function QueueScreen() {
  const { data: channelData } = useChannelList({});

  const { data: nodeData, error: nodeError } = useNodeList({
    visibility: [Visibility.review],
    format: "flat",
  });

  const { data: threadData, error: threadError } = useThreadList({
    visibility: [Visibility.review],
  });

  const isLoading = !nodeData || !threadData || !channelData;
  const error = nodeError || threadError;

  if (isLoading) {
    return <Unready error={error} />;
  }

  const hasNodes = nodeData.nodes && nodeData.nodes.length > 0;
  const hasThreads = threadData.threads && threadData.threads.length > 0;

  const channels = channelData.channels || [];
  const channelMap = new Map(channels.map((ch) => [ch.id, ch]));

  const threadsByChannel = new Map<string, typeof threadData.threads>();
  threadData.threads?.forEach((thread) => {
    if (!threadsByChannel.has(thread.channel_id)) {
      threadsByChannel.set(thread.channel_id, []);
    }
    threadsByChannel.get(thread.channel_id)!.push(thread);
  });

  const sortedChannels = Array.from(threadsByChannel.entries())
    .sort((a, b) => {
      const channelA = channelMap.get(a[0]);
      const channelB = channelMap.get(b[0]);
      if (!channelA || !channelB) return 0;
      return (channelA.name || "").localeCompare(channelB.name || "");
    })
    .map(([channelId, threads]) => ({
      channel: channelMap.get(channelId),
      threads,
    }));

  return (
    <LStack>
      <Heading>Submission queue</Heading>

      <VStack gap="8" w="full">
        {hasThreads && (
          <VStack gap="6" w="full">
            <VStack gap="4" w="full">
              {sortedChannels.map(({ channel, threads }) => (
                <VStack key={channel?.id || "unknown"} gap="3" w="full">
                  <HStack gap="3" alignItems="center">
                    {channel?.icon ? (
                      <styled.img
                        src={getAssetURL(channel.icon.path)}
                        alt={channel.name}
                        w="6"
                        h="6"
                        rounded="md"
                        objectFit="cover"
                        flexShrink="0"
                      />
                    ) : (
                      <styled.div
                        w="6"
                        h="6"
                        rounded="md"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                        flexShrink="0"
                        style={{ background: "#4a9d6f" }}
                      >
                        <MembersIcon width="4" height="4" style={{ color: "#ffffff" }} />
                      </styled.div>
                    )}
                    <Link href={`/channels/${channel?.id}`} style={{ flex: 1 }}>
                      <Heading size="sm" as="h3">
                        {channel?.name || "Unknown Channel"}
                      </Heading>
                    </Link>
                    <styled.span fontSize="xs" color="fg.muted">
                      {threads.length} pending
                    </styled.span>
                  </HStack>
                  <QueueThreadList threads={threads} />
                </VStack>
              ))}
            </VStack>
          </VStack>
        )}

        {hasNodes && (
          <VStack gap="3" w="full">
            <Heading size="sm">Library pages pending review</Heading>
            <QueueNodeList nodes={nodeData.nodes} />
          </VStack>
        )}

        {!hasThreads && !hasNodes && (
          <p>No submissions waiting for review.</p>
        )}
      </VStack>
    </LStack>
  );
}
