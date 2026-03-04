"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { LStack, VStack, styled } from "@/styled-system/jsx";
import { HeaderWithBackArrow } from "@/components/site/Header";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { useThreadList } from "@/api/openapi-client/threads";
import { useAccountGet } from "@/api/openapi-client/accounts";
import { ThreadReference } from "@/api/openapi-schema";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { DocumentTextIcon } from "@heroicons/react/24/outline";
import { Unready } from "@/components/site/Unready";
import { Button } from "@/components/ui/button";

export function MyPostsScreen() {
  const { data: account } = useAccountGet();
  const handle = account?.handle;

  const [currentPage, setCurrentPage] = useState(1);
  const [allThreads, setAllThreads] = useState<ThreadReference[]>([]);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasInitiallyLoaded, setHasInitiallyLoaded] = useState(false);
  const loadedPagesRef = useRef<Set<number>>(new Set());

  const { data: threads, error } = useThreadList(
    handle ? { author: handle, page: String(currentPage) } : undefined,
    { swr: { enabled: !!handle } },
  );

  useEffect(() => {
    const pageNum = threads?.current_page;
    if (!threads?.threads || pageNum === undefined) return;

    setHasInitiallyLoaded(true);

    if (loadedPagesRef.current.has(pageNum)) {
      const updatedThreadsMap = new Map(threads.threads.map((t) => [t.id, t]));
      setAllThreads((prev) =>
        prev.map((t) => updatedThreadsMap.get(t.id) ?? t)
      );
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

  const hasMore = threads?.next_page !== undefined;

  return (
    <LStack gap="0" width="full" height="screen" style={{ backgroundColor: "white" }}>
      <HeaderWithBackArrow
        title="My Posts"
        subtitle="All your community posts"
        headerIcon={<DocumentTextIcon width="20" height="20" style={{ color: "#ffffff" }} />}
        headerIconBackground={`${TRAYA_COLORS.primary}`}
        mobileOnly
        isSticky
      />

      <styled.div
        flex="1"
        width="full"
        style={{
          overflowY: "auto",
          minHeight: 0,
        }}
      >
        <VStack gap="4" width="full" px="4" py="4">
          {!handle ? (
            <Unready />
          ) : error ? (
            <Unready error={error} />
          ) : !hasInitiallyLoaded ? (
            <styled.div p="8" textAlign="center" color="fg.muted" width="full">
              Loading...
            </styled.div>
          ) : allThreads.length === 0 ? (
            <VStack gap="2" py="12" alignItems="center">
              <DocumentTextIcon width="48" height="48" style={{ color: TRAYA_COLORS.neutral.textMuted }} />
              <styled.p fontSize="md" fontWeight="medium" color="fg.muted">
                No posts yet
              </styled.p>
              <styled.p fontSize="sm" color="fg.muted" textAlign="center">
                Your community posts will appear here
              </styled.p>
            </VStack>
          ) : (
            <>
              <VStack alignItems="start" gap="4" width="full">
                {allThreads.map((thread) => (
                  <ThreadReferenceCard key={thread.id} thread={thread} />
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
          )}
        </VStack>
      </styled.div>
    </LStack>
  );
}
