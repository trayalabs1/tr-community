"use client";

import Link from "next/link";
import { Fragment, useState, useEffect, useCallback, useRef } from "react";

import { Unready } from "src/components/site/Unready";
import { LoadingBanner } from "@/components/site/Loading";

import {
  useChannelCategoryList,
  useChannelThreadList,
} from "@/api/openapi-client/channels";
import { Account, Channel, ThreadReference } from "@/api/openapi-schema";
import { ChannelMobileHeader } from "@/components/channel/ChannelMobileHeader";
import { ChannelFilterBar } from "@/components/channel/ChannelFilterBar";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { ShareExperiencePrompt } from "@/components/thread/ThreadCreate/ShareExperiencePrompt";
import { Button } from "@/components/ui/button";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { useChannelPermissions } from "@/lib/channel/permissions";

type Props = {
  session?: Account;
  channel: Channel;
  hasUnreadNotifications?: boolean;
  bookmarkCount?: number;
};

export function ChannelScreen(props: Props) {
  const permissions = useChannelPermissions(props.channel.id);
  const [selectedCategorySlug, setSelectedCategorySlug] = useState<string | null>(null);
  const [selectedVisibility, setSelectedVisibility] = useState<string | null>(null);
  const [dateRange, setDateRange] = useState<{ createdAfter?: string; createdBefore?: string }>({});
  const [excludeBAH, setExcludeBAH] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [allThreads, setAllThreads] = useState<ThreadReference[]>([]);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasInitiallyLoaded, setHasInitiallyLoaded] = useState(false);
  const loadedPagesRef = useRef<Set<number>>(new Set());

  const { data: categories, isValidating: isCategoriesLoading } = useChannelCategoryList(props.channel.id, {
    swr: {
      revalidateOnFocus: false,
    },
  });

  const threadParams: Record<string, string | string[]> = {
    page: String(currentPage),
  };
  if (selectedCategorySlug) {
    threadParams['categories'] = [selectedCategorySlug];
  }
  if (selectedVisibility) {
    threadParams['visibility'] = [selectedVisibility];
  }
  if (dateRange.createdAfter) {
    threadParams['created_after'] = dateRange.createdAfter;
  }
  if (dateRange.createdBefore) {
    threadParams['created_before'] = dateRange.createdBefore;
  }
  if (excludeBAH) {
    threadParams['exclude_bah'] = 'true';
  }

  const { data: threads, error, isValidating: isThreadsLoading } = useChannelThreadList(
    props.channel.id,
    threadParams,
    {
      swr: {
        revalidateOnFocus: false,
      },
    }
  );

  useEffect(() => {
    loadedPagesRef.current = new Set();
    setCurrentPage(1);
    setAllThreads([]);
    setHasInitiallyLoaded(false);
  }, [selectedCategorySlug, selectedVisibility, dateRange, excludeBAH]);

  useEffect(() => {
    const pageNum = threads?.current_page;
    if (!threads?.threads || pageNum === undefined) return;

    setHasInitiallyLoaded(true);

    // If page already loaded, sync updates (for optimistic updates like likes)
    if (loadedPagesRef.current.has(pageNum)) {
      setAllThreads((prev) => {
        const freshById = new Map(threads.threads.map((t) => [t.id, t]));
        const updated = prev.map((t) => freshById.get(t.id) ?? t);
        // For page 1 revalidation, also prepend any newly created threads
        if (pageNum === 1) {
          const existingIds = new Set(prev.map((t) => t.id));
          const newThreads = threads.threads.filter((t) => !existingIds.has(t.id));
          if (newThreads.length > 0) {
            return [...newThreads, ...updated];
          }
        }
        return updated;
      });
      setIsLoadingMore(false);
      return;
    }

    loadedPagesRef.current.add(pageNum);

    if (pageNum === 1) {
      setAllThreads(threads.threads);
    } else {
      setAllThreads((prev) => {
        const existingIds = new Set(prev.map((t) => t.id));
        const newThreads = threads.threads.filter((t) => !existingIds.has(t.id));
        return [...prev, ...newThreads];
      });
    }
    setIsLoadingMore(false);
  }, [threads]);

  const handleLoadMore = useCallback(() => {
    if (threads?.next_page) {
      setIsLoadingMore(true);
      setCurrentPage(threads.next_page);
    }
  }, [threads?.next_page]);

  if (error) {
    return <Unready error={error} />;
  }

  const isInitialLoading = isCategoriesLoading || (isThreadsLoading && allThreads.length === 0);

  if (isInitialLoading && !categories && !threads) {
    return <LoadingBanner />;
  }

  const hasMore = threads?.next_page !== undefined;

  return (
    <LStack gap="0" p="0">
      <ChannelMobileHeader
        channel={props.channel}
        session={props.session}
        categories={categories?.categories || []}
        selectedCategorySlug={selectedCategorySlug}
        selectedVisibility={selectedVisibility}
        onCategoryChange={setSelectedCategorySlug}
        onVisibilityChange={setSelectedVisibility}
        onDateRangeChange={setDateRange}
        excludeBAH={excludeBAH}
        onExcludeBAHChange={setExcludeBAH}
        hasUnreadNotifications={props.hasUnreadNotifications}
        bookmarkCount={props.bookmarkCount}
      />

      <LStack gap="6" p="4">

      {/* Desktop Channel Header */}
      <VStack alignItems="start" gap="2" width="full" display={{ base: "none", md: "flex" }}>
        <HStack justifyContent="space-between" width="full">
          <Heading as="h1" size="2xl">
            {props.channel.name}
          </Heading>
          {permissions.canManageChannel && (
            <Button asChild size="sm" variant="ghost">
              <Link href={`/channels/${props.channel.id}/settings`}>
                Settings
              </Link>
            </Button>
          )}
        </HStack>
        {props.channel.description && (
          <styled.p color="fg.muted">{props.channel.description}</styled.p>
        )}
      </VStack>

      {/* Filter Bar - Desktop Only */}
      <styled.div display={{ base: "none", md: "block" }} width="full">
        <ChannelFilterBar
          channelID={props.channel.id}
          categories={categories?.categories || []}
          selectedCategorySlug={selectedCategorySlug}
          selectedVisibility={selectedVisibility}
          onCategoryChange={setSelectedCategorySlug}
          onVisibilityChange={setSelectedVisibility}
          onDateRangeChange={setDateRange}
          excludeBAH={excludeBAH}
          onExcludeBAHChange={setExcludeBAH}
        />
      </styled.div>

      {/* Threads Section */}
      <VStack alignItems="start" gap="4" width="full">

        {!hasInitiallyLoaded ? (
          <styled.div
            p="8"
            textAlign="center"
            color="fg.muted"
            width="full"
          >
            Loading threads...
          </styled.div>
        ) : allThreads.length > 0 ? (
          <>
            <VStack alignItems="start" gap="4" width="full">
              {allThreads.map((thread, index) => (
                <Fragment key={thread.id}>
                  <ThreadReferenceCard
                    thread={thread}
                    channelID={props.channel.id}
                  />
                  {index === 4 && (
                    <ShareExperiencePrompt channelID={props.channel.id} />
                  )}
                </Fragment>
              ))}
            </VStack>

            {hasMore && (
              <styled.div width="full" display="flex" justifyContent="center" py="4">
                <Button
                  variant="outline"
                  size="md"
                  onClick={handleLoadMore}
                  disabled={isLoadingMore}
                >
                  {isLoadingMore ? "Loading..." : "Show More"}
                </Button>
              </styled.div>
            )}
          </>
        ) : (
          <styled.div
            p="8"
            textAlign="center"
            color="fg.muted"
            borderColor="border.default"
            borderRadius="md"
            borderStyle="solid"
            width="full"
          >
            No threads yet in this channel
          </styled.div>
        )}
      </VStack>
      </LStack>
    </LStack>
  );
}
