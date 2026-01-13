"use client";

import Link from "next/link";
import { useState } from "react";
import { Filter, X } from "lucide-react";
import { useChannelList } from "@/api/openapi-client/channels";
import { useNodeList } from "@/api/openapi-client/nodes";
import { useThreadList } from "@/api/openapi-client/threads";
import { Visibility } from "@/api/openapi-schema";
import { QueueNodeList } from "@/components/queue/QueueNodeList";
import { QueueThreadList } from "@/components/queue/QueueThreadList";
import { LoadingBanner } from "@/components/site/Loading";
import { Heading } from "@/components/ui/heading";
import { VStack, LStack, HStack, styled } from "@/styled-system/jsx";
import { getAssetURL } from "@/utils/asset";
import { MembersIcon } from "@/components/ui/icons/Members";

type ContentType = "threads" | "nodes" | "all";

export function QueueScreen() {
  const [selectedChannelId, setSelectedChannelId] = useState<string | null>(null);
  const [selectedContentType, setSelectedContentType] = useState<ContentType>("all");
  const [isFiltersOpen, setIsFiltersOpen] = useState(false);

  const { data: channelData } = useChannelList({});

  const { data: nodeData, isValidating: isNodesLoading } = useNodeList({
    visibility: [Visibility.review],
    format: "flat",
  });

  const { data: threadData, isValidating: isThreadsLoading } = useThreadList({
    visibility: [Visibility.review],
  });

  const isLoading = isThreadsLoading && isNodesLoading;

  if (isLoading && !nodeData && !threadData) {
    return <LoadingBanner />;
  }

  const hasNodes = nodeData?.nodes && nodeData.nodes.length > 0;
  const hasThreads = threadData?.threads && threadData.threads.length > 0;

  const channels = channelData?.channels || [];
  const channelMap = new Map(channels.map((ch) => [ch.id, ch]));

  const threadsByChannel = new Map<string, any[]>();
  threadData?.threads?.forEach((thread) => {
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

  const filteredChannels = selectedChannelId
    ? sortedChannels.filter(({ channel }) => channel?.id === selectedChannelId)
    : sortedChannels;

  const shouldShowThreads = selectedContentType === "threads" || selectedContentType === "all";
  const shouldShowNodes = selectedContentType === "nodes" || selectedContentType === "all";

  const totalPending = (threadData?.threads?.length || 0) + (nodeData?.nodes?.length || 0);

  return (
    <LStack gap="6" p="4">
      {/* Header */}
      <VStack alignItems="start" gap="2" width="full">
        <HStack justifyContent="space-between" width="full" alignItems="center">
          <VStack alignItems="start" gap="1">
            <Heading as="h1" size="2xl">
              Submission Queue
            </Heading>
            <styled.p fontSize="sm" color="fg.muted">
              {totalPending} {totalPending === 1 ? "submission" : "submissions"} pending review
            </styled.p>
          </VStack>
          <styled.button
            onClick={() => setIsFiltersOpen(!isFiltersOpen)}
            display="flex"
            alignItems="center"
            gap="2"
            fontSize="sm"
            fontWeight="semibold"
            cursor="pointer"
            px="3"
            py="2"
            rounded="md"
            style={{
              backgroundColor: isFiltersOpen ? "var(--colors-bg-muted)" : "transparent",
              color: "var(--colors-fg-default)",
              border: "1px solid var(--colors-border-default)",
              transition: "all 0.2s ease-in-out",
            }}
          >
            <Filter size={18} strokeWidth={2} />
            <styled.span>Filters</styled.span>
          </styled.button>
        </HStack>
      </VStack>

      {/* Filters Panel */}
      {isFiltersOpen && (
        <VStack
          alignItems="start"
          gap="4"
          width="full"
          p="4"
          rounded="md"
          style={{
            border: "1px solid var(--colors-border-default)",
            backgroundColor: "var(--colors-bg-muted)",
          }}
        >
          {/* Content Type Filter */}
          <VStack alignItems="start" gap="2" width="full">
            <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
              Content Type
            </styled.label>
            <HStack gap="3" flexWrap="wrap">
              {(["all", "threads", "nodes"] as const).map((type) => (
                <styled.button
                  key={type}
                  onClick={() => setSelectedContentType(type)}
                  fontSize="sm"
                  cursor="pointer"
                  transition="all"
                  style={{
                    backgroundColor: "transparent",
                    color: selectedContentType === type ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
                    border: "none",
                    borderBottom: selectedContentType === type ? "2px solid var(--colors-fg-default)" : "2px solid transparent",
                    padding: "4px 0",
                    fontWeight: selectedContentType === type ? "600" : "400",
                  }}
                >
                  {type === "all" ? "All" : type === "threads" ? "Threads" : "Library Pages"}
                </styled.button>
              ))}
            </HStack>
          </VStack>

          {/* Channel Filter */}
          {hasThreads && (
            <VStack alignItems="start" gap="2" width="full">
              <HStack justifyContent="space-between" width="full">
                <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
                  Channel
                </styled.label>
                {selectedChannelId && (
                  <styled.button
                    onClick={() => setSelectedChannelId(null)}
                    fontSize="xs"
                    color="fg.muted"
                    cursor="pointer"
                    style={{
                      background: "none",
                      border: "none",
                      padding: "0",
                      textDecoration: "underline",
                    }}
                  >
                    Clear
                  </styled.button>
                )}
              </HStack>
              <HStack gap="2" flexWrap="wrap">
                {sortedChannels.map(({ channel }) => (
                  <styled.button
                    key={channel?.id}
                    onClick={() =>
                      setSelectedChannelId(selectedChannelId === channel?.id ? null : channel?.id || null)
                    }
                    fontSize="xs"
                    cursor="pointer"
                    px="3"
                    py="1.5"
                    rounded="full"
                    transition="all"
                    style={{
                      backgroundColor:
                        selectedChannelId === channel?.id
                          ? "var(--colors-bg-default)"
                          : "var(--colors-bg-subtle)",
                      color:
                        selectedChannelId === channel?.id
                          ? "var(--colors-fg-default)"
                          : "var(--colors-fg-muted)",
                      border: "1px solid var(--colors-border-default)",
                      fontWeight: selectedChannelId === channel?.id ? "600" : "400",
                    }}
                  >
                    {channel?.name || "Unknown"}
                  </styled.button>
                ))}
              </HStack>
            </VStack>
          )}
        </VStack>
      )}

      {/* Content */}
      <VStack gap="8" width="full">
        {/* Threads Section */}
        {shouldShowThreads && hasThreads && (
          <VStack gap="6" width="full">
            <HStack justifyContent="space-between" width="full" alignItems="center">
              <Heading as="h2" size="lg">
                Threads Pending Review
              </Heading>
              <styled.span fontSize="xs" color="fg.muted" fontWeight="semibold">
                {filteredChannels.reduce((sum, { threads }) => sum + threads.length, 0)} pending
              </styled.span>
            </HStack>

            {filteredChannels.length > 0 ? (
              <VStack gap="4" width="full">
                {filteredChannels.map(({ channel, threads }) => (
                  <VStack key={channel?.id || "unknown"} gap="3" width="full">
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
            ) : (
              <styled.div
                p="8"
                textAlign="center"
                color="fg.muted"
                width="full"
                style={{
                  border: "1px solid var(--colors-border-subtle)",
                  borderRadius: "0.5rem",
                  backgroundColor: "var(--colors-bg-muted)",
                }}
              >
                No threads pending review in selected channel
              </styled.div>
            )}
          </VStack>
        )}

        {/* Library Pages Section */}
        {shouldShowNodes && hasNodes && (
          <VStack gap="3" width="full">
            <HStack justifyContent="space-between" width="full" alignItems="center">
              <Heading as="h2" size="lg">
                Library Pages Pending Review
              </Heading>
              <styled.span fontSize="xs" color="fg.muted" fontWeight="semibold">
                {nodeData?.nodes?.length || 0} pending
              </styled.span>
            </HStack>
            <QueueNodeList nodes={nodeData?.nodes || []} />
          </VStack>
        )}

        {/* Empty State */}
        {!hasThreads && !hasNodes && (
          <VStack
            gap="4"
            width="full"
            p="12"
            alignItems="center"
            justifyContent="center"
            style={{
              border: "1px dashed var(--colors-border-default)",
              borderRadius: "0.75rem",
              backgroundColor: "var(--colors-bg-muted)",
              minHeight: "300px",
            }}
          >
            <styled.div fontSize="lg" fontWeight="semibold" color="fg.muted">
              All Caught Up!
            </styled.div>
            <styled.p fontSize="sm" color="fg.muted">
              No submissions waiting for review at the moment.
            </styled.p>
          </VStack>
        )}
      </VStack>
    </LStack>
  );
}
