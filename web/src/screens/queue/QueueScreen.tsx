"use client";

import Link from "next/link";
import { useState, useMemo, useCallback, useEffect, useRef } from "react";
import { Filter } from "lucide-react";
import { today, getLocalTimeZone, type DateValue } from "@internationalized/date";
import { DateRangePicker } from "@/components/ui/date-picker";
import { useAdminReplyQueueList, adminReplyQueueDismiss } from "@/api/openapi-client/admin";
import { useChannelList } from "@/api/openapi-client/channels";
import { useNodeList } from "@/api/openapi-client/nodes";
import { useThreadList } from "@/api/openapi-client/threads";
import { Visibility, ThreadReference, ReplyQueueEntry } from "@/api/openapi-schema";
import { PendingReplyThreadList } from "@/components/queue/PendingReplyThreadList";
import { QueueNodeList } from "@/components/queue/QueueNodeList";
import { QueueThreadList } from "@/components/queue/QueueThreadList";
import { ReplyAdminQueueList } from "@/components/queue/ReplyAdminQueueList";
import { LoadingBanner } from "@/components/site/Loading";
import { Button } from "@/components/ui/button";
import { Heading } from "@/components/ui/heading";
import { VStack, LStack, HStack, styled } from "@/styled-system/jsx";
import { getAssetURL } from "@/utils/asset";
import { MembersIcon } from "@/components/ui/icons/Members";

type ContentType = "threads" | "nodes" | "all";
type QueueTab = "pending_review" | "pending_reply" | "pending_reply_to_reply";

export function QueueScreen() {
  const [activeTab, setActiveTab] = useState<QueueTab>("pending_review");
  const [selectedChannelId, setSelectedChannelId] = useState<string | null>(null);
  const [selectedContentType, setSelectedContentType] = useState<ContentType>("all");
  const [isFiltersOpen, setIsFiltersOpen] = useState(false);
  const [excludeBAH, setExcludeBAH] = useState(true);

  const todayVal = useMemo(() => today(getLocalTimeZone()), []);

  const [pendingReplyRange, setPendingReplyRange] = useState<{
    createdAfter: string;
    createdBefore: string;
  }>(() => {
    const tz = getLocalTimeZone();
    const start = todayVal.subtract({ days: 2 }).toDate(tz);
    start.setHours(0, 0, 0, 0);
    const end = todayVal.toDate(tz);
    end.setHours(23, 59, 59, 999);
    return { createdAfter: start.toISOString(), createdBefore: end.toISOString() };
  });

  const [pendingReplyPage, setPendingReplyPage] = useState(1);
  const [allPendingReplyThreads, setAllPendingReplyThreads] = useState<ThreadReference[]>([]);
  const [isPendingReplyLoadingMore, setIsPendingReplyLoadingMore] = useState(false);
  const pendingReplyLoadedPages = useRef<Set<number>>(new Set());

  const [replyQueuePage, setReplyQueuePage] = useState(1);
  const [allReplyQueueEntries, setAllReplyQueueEntries] = useState<ReplyQueueEntry[]>([]);
  const [isReplyQueueLoadingMore, setIsReplyQueueLoadingMore] = useState(false);
  const replyQueueLoadedPages = useRef<Set<number>>(new Set());

  const { data: channelData } = useChannelList({});

  const { data: nodeData, isValidating: isNodesLoading } = useNodeList({
    visibility: [Visibility.review],
    format: "flat",
  });

  const { data: threadData, isValidating: isThreadsLoading } = useThreadList({
    visibility: [Visibility.review],
    ...(excludeBAH && { exclude_bah: true }),
  });

  const { data: pendingReplyData, isValidating: isPendingReplyLoading } = useThreadList({
    visibility: [Visibility.published],
    created_after: pendingReplyRange.createdAfter,
    created_before: pendingReplyRange.createdBefore,
    no_replies: true,
    ...(excludeBAH && { exclude_bah: true }),
    page: String(pendingReplyPage),
  });

  const { data: replyQueueData, mutate: mutateReplyQueue, isValidating: isReplyQueueLoading } = useAdminReplyQueueList({
    created_after: pendingReplyRange.createdAfter,
    created_before: pendingReplyRange.createdBefore,
    page: String(replyQueuePage),
  });

  useEffect(() => {
    const pageNum = pendingReplyData?.current_page;
    if (!pendingReplyData?.threads || pageNum === undefined) return;

    if (pendingReplyLoadedPages.current.has(pageNum)) {
      setIsPendingReplyLoadingMore(false);
      return;
    }
    pendingReplyLoadedPages.current.add(pageNum);

    if (pageNum === 1) {
      setAllPendingReplyThreads(pendingReplyData.threads);
    } else {
      setAllPendingReplyThreads((prev) => {
        const existingIds = new Set(prev.map((t) => t.id));
        const newThreads = pendingReplyData.threads.filter((t) => !existingIds.has(t.id));
        return [...prev, ...newThreads];
      });
    }
    setIsPendingReplyLoadingMore(false);
  }, [pendingReplyData]);

  useEffect(() => {
    const pageNum = replyQueueData?.current_page;
    if (!replyQueueData?.reply_queue_entries || pageNum === undefined) return;

    if (replyQueueLoadedPages.current.has(pageNum)) {
      setIsReplyQueueLoadingMore(false);
      return;
    }
    replyQueueLoadedPages.current.add(pageNum);

    if (pageNum === 1) {
      setAllReplyQueueEntries(replyQueueData.reply_queue_entries);
    } else {
      setAllReplyQueueEntries((prev) => {
        const existingIds = new Set(prev.map((e) => e.id));
        const newEntries = replyQueueData.reply_queue_entries.filter((e) => !existingIds.has(e.id));
        return [...prev, ...newEntries];
      });
    }
    setIsReplyQueueLoadingMore(false);
  }, [replyQueueData]);

  const handleLoadMorePendingReply = useCallback(() => {
    if (pendingReplyData?.next_page) {
      setIsPendingReplyLoadingMore(true);
      setPendingReplyPage(pendingReplyData.next_page);
    }
  }, [pendingReplyData?.next_page]);

  const handleLoadMoreReplyQueue = useCallback(() => {
    if (replyQueueData?.next_page) {
      setIsReplyQueueLoadingMore(true);
      setReplyQueuePage(replyQueueData.next_page);
    }
  }, [replyQueueData?.next_page]);

  const resetPagination = useCallback(() => {
    setPendingReplyPage(1);
    setAllPendingReplyThreads([]);
    pendingReplyLoadedPages.current = new Set();
    setReplyQueuePage(1);
    setAllReplyQueueEntries([]);
    replyQueueLoadedPages.current = new Set();
  }, []);

  const handlePendingReplyDateChange = useCallback(({ value }: { value: DateValue[] }) => {
    const [start, end] = value;

    if (!start) {
      const tz = getLocalTimeZone();
      const s = todayVal.subtract({ days: 2 }).toDate(tz);
      s.setHours(0, 0, 0, 0);
      const e = todayVal.toDate(tz);
      e.setHours(23, 59, 59, 999);
      setPendingReplyRange({ createdAfter: s.toISOString(), createdBefore: e.toISOString() });
      resetPagination();
      return;
    }

    if (!end) return;

    const [earlier, later] = start.compare(end) <= 0 ? [start, end] : [end, start];
    const effectiveEnd = later.compare(earlier) > 2 ? earlier.add({ days: 2 }) : later;

    const startDate = earlier.toDate(getLocalTimeZone());
    startDate.setHours(0, 0, 0, 0);
    const endOfDay = new Date(effectiveEnd.add({ days: 1 }).toDate(getLocalTimeZone()).getTime() - 1);

    setPendingReplyRange({
      createdAfter: startDate.toISOString(),
      createdBefore: endOfDay.toISOString(),
    });
    resetPagination();
  }, [todayVal, resetPagination]);

  const isInitialLoading =
    activeTab === "pending_review" ? isThreadsLoading && isNodesLoading :
    activeTab === "pending_reply" ? isPendingReplyLoading && allPendingReplyThreads.length === 0 :
    isReplyQueueLoading && allReplyQueueEntries.length === 0;

  if (isInitialLoading) {
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

  const pendingReplyThreads = allPendingReplyThreads;
  const replyQueueEntries = allReplyQueueEntries;

  const allChannelIds = new Set([
    ...Array.from(threadsByChannel.keys()),
    ...pendingReplyThreads.map((t) => t.channel_id).filter((id): id is string => !!id),
    ...replyQueueEntries.map((e) => e.channel_id).filter((id): id is string => !!id),
  ]);
  const allChannelOptions = Array.from(allChannelIds)
    .map((id) => channelMap.get(id))
    .filter((ch): ch is NonNullable<typeof ch> => ch != null)
    .sort((a, b) => (a.name || "").localeCompare(b.name || ""));

  const filteredPendingReplyThreads = selectedChannelId
    ? pendingReplyThreads.filter((t) => t.channel_id === selectedChannelId)
    : pendingReplyThreads;

  return (
    <LStack gap="6" p="4">
      {/* Header */}
      <VStack alignItems="start" gap="2" width="full">
        <HStack justifyContent="space-between" width="full" alignItems="center">
          <Heading as="h1" size="2xl">
            Submission Queue
          </Heading>
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

        {/* Tabs */}
        <HStack gap="0" width="full" style={{ borderBottom: "1px solid var(--colors-border-default)" }}>
          {(["pending_review", "pending_reply", "pending_reply_to_reply"] as const).map((tab) => (
            <styled.button
              key={tab}
              onClick={() => setActiveTab(tab)}
              fontSize="sm"
              cursor="pointer"
              px="4"
              py="2"
              style={{
                background: "none",
                border: "none",
                borderBottom: activeTab === tab ? "2px solid var(--colors-fg-default)" : "2px solid transparent",
                marginBottom: "-1px",
                fontWeight: activeTab === tab ? "600" : "400",
                color: activeTab === tab ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
              }}
            >
              {tab === "pending_review" ? "Pending Review" : tab === "pending_reply" ? "Pending Reply" : "Pending Reply to Reply"}
            </styled.button>
          ))}
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

          {allChannelOptions.length > 0 && (
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
                {allChannelOptions.map((channel) => (
                  <styled.button
                    key={channel.id}
                    onClick={() =>
                      setSelectedChannelId(selectedChannelId === channel.id ? null : channel.id)
                    }
                    fontSize="xs"
                    cursor="pointer"
                    px="3"
                    py="1.5"
                    rounded="full"
                    transition="all"
                    style={{
                      backgroundColor:
                        selectedChannelId === channel.id
                          ? "var(--colors-bg-default)"
                          : "var(--colors-bg-subtle)",
                      color:
                        selectedChannelId === channel.id
                          ? "var(--colors-fg-default)"
                          : "var(--colors-fg-muted)",
                      border: "1px solid var(--colors-border-default)",
                      fontWeight: selectedChannelId === channel.id ? "600" : "400",
                    }}
                  >
                    {channel.name || "Unknown"}
                  </styled.button>
                ))}
              </HStack>
            </VStack>
          )}

          {(activeTab === "pending_reply" || activeTab === "pending_reply_to_reply") && (
            <VStack alignItems="start" gap="2" width="full">
              <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
                Date Range
              </styled.label>
              <DateRangePicker
                hideInputs={true}
                min={todayVal.subtract({ days: 3 })}
                max={todayVal}
                onValueChange={handlePendingReplyDateChange}
              />
            </VStack>
          )}

          <VStack alignItems="start" gap="2" width="full">
            <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
              Streak Posts
            </styled.label>
            <HStack gap="2" flexWrap="wrap">
              <styled.button
                onClick={() => setExcludeBAH(!excludeBAH)}
                fontSize="xs"
                cursor="pointer"
                px="3"
                py="1.5"
                rounded="full"
                transition="all"
                style={{
                  backgroundColor: excludeBAH ? "var(--colors-bg-default)" : "var(--colors-bg-subtle)",
                  color: excludeBAH ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
                  border: "1px solid var(--colors-border-default)",
                  fontWeight: excludeBAH ? "600" : "400",
                }}
              >
                Hide Streak Posts
              </styled.button>
            </HStack>
          </VStack>
        </VStack>
      )}

      {/* Pending Reply Tab */}
      {activeTab === "pending_reply" && (
        <VStack gap="4" width="full">
          <HStack justifyContent="space-between" width="full" alignItems="center">
            <Heading as="h2" size="lg">
              Threads Pending Reply
            </Heading>
            <styled.span fontSize="xs" color="fg.muted" fontWeight="semibold">
              {filteredPendingReplyThreads.length} pending
            </styled.span>
          </HStack>
          {shouldShowThreads && filteredPendingReplyThreads.length > 0 ? (
            <>
              <PendingReplyThreadList threads={filteredPendingReplyThreads} />
              {pendingReplyData?.next_page && (
                <styled.div width="full" display="flex" justifyContent="center" py="4">
                  <Button
                    variant="outline"
                    size="md"
                    onClick={handleLoadMorePendingReply}
                    disabled={isPendingReplyLoadingMore}
                  >
                    {isPendingReplyLoadingMore ? "Loading..." : "Load More"}
                  </Button>
                </styled.div>
              )}
            </>
          ) : (
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
                All Replied!
              </styled.div>
              <styled.p fontSize="sm" color="fg.muted">
                No threads pending reply for the selected filters.
              </styled.p>
            </VStack>
          )}
        </VStack>
      )}

      {/* Pending Reply to Reply Tab */}
      {activeTab === "pending_reply_to_reply" && (
        <VStack gap="4" width="full">
          <HStack justifyContent="space-between" width="full" alignItems="center">
            <Heading as="h2" size="lg">
              Replies Pending Admin Attention
            </Heading>
            <styled.span fontSize="xs" color="fg.muted" fontWeight="semibold">
              {replyQueueEntries.length} pending
            </styled.span>
          </HStack>
          <ReplyAdminQueueList
            entries={selectedChannelId
              ? replyQueueEntries.filter((e) => e.channel_id === selectedChannelId)
              : replyQueueEntries}
            channelMap={channelMap}
            onDismiss={async (id) => {
              await adminReplyQueueDismiss(id);
              setAllReplyQueueEntries((prev) => prev.filter((e) => e.id !== id));
              await mutateReplyQueue();
            }}
          />
          {replyQueueData?.next_page && (
            <styled.div width="full" display="flex" justifyContent="center" py="4">
              <Button
                variant="outline"
                size="md"
                onClick={handleLoadMoreReplyQueue}
                disabled={isReplyQueueLoadingMore}
              >
                {isReplyQueueLoadingMore ? "Loading..." : "Load More"}
              </Button>
            </styled.div>
          )}
        </VStack>
      )}

      {/* Pending Review Tab */}
      {activeTab === "pending_review" && <VStack gap="8" width="full">
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
      </VStack>}
    </LStack>
  );
}
