import { useCallback, useEffect, useMemo, useState } from "react";
import { today, getLocalTimeZone, type DateValue } from "@internationalized/date";

import { threadList } from "@/api/openapi-client/threads";
import { likePostAddMany } from "@/api/openapi-client/likes";
import { replyCreateMany } from "@/api/openapi-client/replies";
import { ThreadReference, Visibility } from "@/api/openapi-schema";
import { PendingReplyThreadList } from "@/components/queue/PendingReplyThreadList";
import { DateRangePicker } from "@/components/ui/date-picker";
import { Button } from "@/components/ui/button";
import { Heading } from "@/components/ui/heading";
import { Spinner } from "@/components/ui/Spinner";
import { VStack, HStack, styled } from "@/styled-system/jsx";

const BULK_REPLY_RESPONSES: string[] = [
  "Consistency like this is powerful. You're on the right track keep going 💪",
  "Love the discipline 🔥 This is exactly how results start showing.",
  "Showing up daily like this? That's a big win 💪 Keep it going!",
  "This kind of consistency is rare 👏 You're doing it right.",
  "Great job staying consistent 🔥 Trust the process it's working.",
];

const CATEGORY_OPTIONS: { value: string; label: string }[] = [
  { value: "BAH", label: "Streak Posts" },
  { value: "feedback", label: "Feedback" },
];

const SENTIMENT_OPTIONS: { value: string; label: string }[] = [
  { value: "positive", label: "Positive" },
  { value: "neutral", label: "Neutral" },
  { value: "negative", label: "Negative" },
];

function chipStyle(active: boolean) {
  return {
    backgroundColor: active ? "var(--colors-bg-default)" : "var(--colors-bg-subtle)",
    color: active ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
    border: "1px solid var(--colors-border-default)",
    fontWeight: active ? "600" : "400",
  };
}

function ChipFilter({
  label,
  options,
  isActive,
  onSelect,
}: {
  label: string;
  options: { value: string; label: string }[];
  isActive: (value: string) => boolean;
  onSelect: (value: string) => void;
}) {
  return (
    <VStack alignItems="start" gap="2" width="full">
      <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
        {label}
      </styled.label>
      <HStack gap="2" flexWrap="wrap">
        {options.map((opt) => (
          <styled.button
            key={opt.value}
            onClick={() => onSelect(opt.value)}
            fontSize="xs"
            cursor="pointer"
            px="3"
            py="1.5"
            rounded="full"
            transition="all"
            style={chipStyle(isActive(opt.value))}
          >
            {opt.label}
          </styled.button>
        ))}
      </HStack>
    </VStack>
  );
}

// Even split: divide threads into N contiguous groups (N = number of replies)
// and assign all threads in a group the same reply. The first groups absorb the
// remainder when the counts don't divide evenly.
function assignRepliesEvenly(threadCount: number, replies: string[]): string[] {
  if (replies.length === 0 || threadCount === 0) return [];

  const n = replies.length;
  const base = Math.floor(threadCount / n);
  const remainder = threadCount % n;

  const assignments: string[] = [];
  for (let r = 0; r < n; r++) {
    const groupSize = base + (r < remainder ? 1 : 0);
    for (let k = 0; k < groupSize; k++) {
      assignments.push(replies[r]!);
    }
  }
  return assignments;
}

type PendingAction = "like" | "reply" | null;

type Filters = {
  createdAfter: string;
  createdBefore: string;
  categories: string[];
  sentiments: string[];
  noLikes: boolean;
};

async function fetchAllMatchingThreads(filters: Filters): Promise<ThreadReference[]> {
  const all: ThreadReference[] = [];
  let page = 1;
  // Hard cap to avoid runaway loops on unexpectedly large result sets.
  for (let i = 0; i < 100; i++) {
    const result = await threadList({
      visibility: [Visibility.published, Visibility.review],
      created_after: filters.createdAfter,
      created_before: filters.createdBefore,
      no_replies: true,
      ...(filters.noLikes && { no_likes: true }),
      post_categories: filters.categories,
      ...(filters.sentiments.length > 0 && { sentiments: filters.sentiments }),
      page: String(page),
    });

    all.push(...result.threads);

    if (!result.next_page || result.next_page === page) break;
    page = result.next_page;
  }
  return all;
}

function formatRangeLabel(startMs: number, endMs: number): string {
  const fmt = (ms: number) =>
    new Date(ms).toLocaleDateString(undefined, {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  const start = fmt(startMs);
  const end = fmt(endMs);
  return start === end ? start : `${start} – ${end}`;
}

export function BulkActionsPanel() {
  const todayVal = useMemo(() => today(getLocalTimeZone()), []);

  const initialRange = useMemo(() => {
    const tz = getLocalTimeZone();
    const start = todayVal.toDate(tz);
    start.setHours(0, 0, 0, 0);
    const end = todayVal.toDate(tz);
    end.setHours(23, 59, 59, 999);
    return { createdAfter: start.toISOString(), createdBefore: end.toISOString() };
  }, [todayVal]);

  const [range, setRange] = useState(initialRange);
  const [selectedCategory, setSelectedCategory] = useState<string>("BAH");
  const [selectedSentiments, setSelectedSentiments] = useState<Set<string>>(
    () => new Set(),
  );
  const [noLikesOnly, setNoLikesOnly] = useState(false);
  const [selectedReplies, setSelectedReplies] = useState<Set<number>>(
    () => new Set(BULK_REPLY_RESPONSES.map((_, i) => i)),
  );

  const [threads, setThreads] = useState<ThreadReference[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [pendingAction, setPendingAction] = useState<PendingAction>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [resultMessage, setResultMessage] = useState<string | null>(null);

  const rangeLabel = formatRangeLabel(
    Date.parse(range.createdAfter),
    Date.parse(range.createdBefore),
  );

  const hasCategory = selectedCategory !== "";
  // The "No likes" filter forces a sentiment selection so the admin scopes the
  // unliked set before acting on it.
  const sentimentRequired = noLikesOnly && selectedSentiments.size === 0;

  const buildFilters = useCallback(
    (r: { createdAfter: string; createdBefore: string }): Filters => ({
      createdAfter: r.createdAfter,
      createdBefore: r.createdBefore,
      categories: selectedCategory ? [selectedCategory] : [],
      sentiments: Array.from(selectedSentiments),
      noLikes: noLikesOnly,
    }),
    [selectedCategory, selectedSentiments, noLikesOnly],
  );

  const loadThreads = useCallback(
    async (filters: Filters) => {
      if (filters.categories.length === 0) {
        setThreads([]);
        return;
      }
      // No-likes requires a sentiment; don't load an unscoped set.
      if (filters.noLikes && filters.sentiments.length === 0) {
        setThreads([]);
        return;
      }
      setIsLoading(true);
      setResultMessage(null);
      try {
        const result = await fetchAllMatchingThreads(filters);
        setThreads(result);
      } finally {
        setIsLoading(false);
      }
    },
    [],
  );

  // Reload whenever filters change (categories, sentiments, or date range).
  useEffect(() => {
    setPendingAction(null);
    void loadThreads(buildFilters(range));
  }, [buildFilters, loadThreads, range]);

  const selectCategory = useCallback((value: string) => {
    // Single-select: clicking the active category clears it (empties the gate).
    setSelectedCategory((prev) => (prev === value ? "" : value));
  }, []);

  const toggleSentiment = useCallback((value: string) => {
    setSelectedSentiments((prev) => {
      const next = new Set(prev);
      if (next.has(value)) next.delete(value);
      else next.add(value);
      return next;
    });
  }, []);

  const toggleNoLikes = useCallback(() => {
    setNoLikesOnly((prev) => !prev);
  }, []);

  const toggleReply = useCallback((index: number) => {
    setSelectedReplies((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  }, []);

  const handleDateChange = useCallback(
    ({ value }: { value: DateValue[] }) => {
      const [start, end] = value;
      const tz = getLocalTimeZone();

      let nextRange: { createdAfter: string; createdBefore: string };
      if (!start || !end) {
        nextRange = initialRange;
      } else {
        const [earlier, later] = start.compare(end) <= 0 ? [start, end] : [end, start];
        const startDate = earlier.toDate(tz);
        startDate.setHours(0, 0, 0, 0);
        const endOfDay = new Date(later.add({ days: 1 }).toDate(tz).getTime() - 1);
        nextRange = {
          createdAfter: startDate.toISOString(),
          createdBefore: endOfDay.toISOString(),
        };
      }

      setRange(nextRange);
    },
    [initialRange],
  );

  const runBulkAction = useCallback(async () => {
    if (!pendingAction || threads.length === 0) {
      setPendingAction(null);
      return;
    }

    const action = pendingAction;
    setIsRunning(true);

    if (action === "like") {
      // Single request — the server likes every post in one call.
      let liked = 0;
      try {
        const result = await likePostAddMany({ post_ids: threads.map((t) => t.id) });
        liked = result.liked;
      } catch {
        liked = 0;
      }
      setIsRunning(false);
      setPendingAction(null);
      setResultMessage(`Liked ${liked} of ${threads.length} threads.`);
      return;
    }

    // Reply: a single request carrying per-thread bodies, with the selected
    // replies split evenly across the set.
    const replyPool = BULK_REPLY_RESPONSES.filter((_, i) => selectedReplies.has(i));
    const assignments = assignRepliesEvenly(threads.length, replyPool);
    const items = threads
      .map((thread, i) => ({ thread_mark: thread.slug, body: assignments[i] }))
      .filter((item): item is { thread_mark: string; body: string } =>
        Boolean(item.thread_mark && item.body),
      );

    let created = 0;
    try {
      const result = await replyCreateMany({ items });
      created = result.created;
    } catch {
      created = 0;
    }

    setIsRunning(false);
    setPendingAction(null);
    setResultMessage(`Replied to ${created} of ${threads.length} threads.`);

    // Replied threads no longer have zero replies — refresh the matched set.
    void loadThreads(buildFilters(range));
  }, [pendingAction, threads, selectedReplies, loadThreads, buildFilters, range]);

  return (
    <VStack gap="4" width="full" alignItems="start">
      <HStack justifyContent="space-between" width="full" alignItems="center">
        <Heading as="h2" size="lg">
          Bulk Actions
        </Heading>
        {hasCategory && (
          <styled.span fontSize="xs" color="fg.muted" fontWeight="semibold">
            {isLoading ? "Loading…" : `${threads.length} matching`}
          </styled.span>
        )}
      </HStack>

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
        <VStack alignItems="start" gap="2" width="full">
          <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
            Date Range
          </styled.label>
          <HStack gap="3" alignItems="center" flexWrap="wrap">
            <DateRangePicker
              hideInputs={true}
              min={todayVal.subtract({ days: 30 })}
              max={todayVal}
              onValueChange={handleDateChange}
            />
            <styled.span fontSize="sm" fontWeight="medium" color="fg.default">
              {rangeLabel}
            </styled.span>
          </HStack>
        </VStack>

        <ChipFilter
          label="Category"
          options={CATEGORY_OPTIONS}
          isActive={(value) => selectedCategory === value}
          onSelect={selectCategory}
        />

        <ChipFilter
          label="Sentiment"
          options={SENTIMENT_OPTIONS}
          isActive={(value) => selectedSentiments.has(value)}
          onSelect={toggleSentiment}
        />

        <VStack alignItems="start" gap="2" width="full">
          <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
            Engagement
          </styled.label>
          <HStack gap="2" flexWrap="wrap" alignItems="center">
            <styled.button
              onClick={toggleNoLikes}
              fontSize="xs"
              cursor="pointer"
              px="3"
              py="1.5"
              rounded="full"
              transition="all"
              style={chipStyle(noLikesOnly)}
            >
              No likes
            </styled.button>
            {sentimentRequired && (
              <styled.span fontSize="xs" color="fg.error">
                Select a sentiment to use the No likes filter.
              </styled.span>
            )}
          </HStack>
        </VStack>

        <VStack alignItems="start" gap="2" width="full">
          <HStack justifyContent="space-between" width="full">
            <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase">
              Replies for bulk reply
            </styled.label>
            <styled.span fontSize="xs" color="fg.muted">
              {selectedReplies.size} selected — split evenly across threads
            </styled.span>
          </HStack>
          <VStack alignItems="start" gap="1" width="full">
            {BULK_REPLY_RESPONSES.map((response, index) => {
              const isChecked = selectedReplies.has(index);
              return (
                <styled.label
                  key={index}
                  display="flex"
                  gap="2"
                  alignItems="start"
                  fontSize="sm"
                  cursor="pointer"
                  width="full"
                  py="1"
                >
                  <styled.input
                    type="checkbox"
                    checked={isChecked}
                    onChange={() => toggleReply(index)}
                    mt="1"
                    flexShrink="0"
                  />
                  <styled.span color={isChecked ? "fg.default" : "fg.muted"}>
                    {response}
                  </styled.span>
                </styled.label>
              );
            })}
          </VStack>
        </VStack>
      </VStack>

      {!hasCategory || sentimentRequired ? (
        <VStack
          gap="2"
          width="full"
          p="12"
          alignItems="center"
          justifyContent="center"
          style={{
            border: "1px dashed var(--colors-border-default)",
            borderRadius: "0.75rem",
            backgroundColor: "var(--colors-bg-muted)",
            minHeight: "200px",
          }}
        >
          <styled.p fontSize="sm" color="fg.muted">
            {!hasCategory
              ? "Select a category to load threads for bulk actions."
              : "Select a sentiment to use the No likes filter."}
          </styled.p>
        </VStack>
      ) : (
        <>
          {pendingAction ? (
            <HStack
              gap="3"
              width="full"
              p="4"
              rounded="md"
              alignItems="center"
              justifyContent="space-between"
              style={{
                border: "1px solid var(--colors-border-default)",
                backgroundColor: "var(--colors-bg-default)",
              }}
            >
              <styled.span fontSize="sm" fontWeight="semibold">
                {pendingAction === "like"
                  ? `Like ${threads.length} thread${threads.length === 1 ? "" : "s"}?`
                  : `Reply to ${threads.length} thread${threads.length === 1 ? "" : "s"} using ${selectedReplies.size} repl${selectedReplies.size === 1 ? "y" : "ies"}, split evenly?`}
              </styled.span>
              <HStack gap="2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPendingAction(null)}
                  disabled={isRunning}
                >
                  Cancel
                </Button>
                <Button size="sm" onClick={runBulkAction} disabled={isRunning}>
                  {isRunning ? "Running…" : "Confirm"}
                </Button>
              </HStack>
            </HStack>
          ) : (
            <HStack gap="2" width="full">
              <Button
                variant="outline"
                size="md"
                onClick={() => setPendingAction("like")}
                disabled={isLoading || threads.length === 0}
              >
                Bulk Like
              </Button>
              <Button
                size="md"
                onClick={() => setPendingAction("reply")}
                disabled={
                  isLoading ||
                  threads.length === 0 ||
                  selectedReplies.size === 0 ||
                  noLikesOnly
                }
              >
                Bulk Reply
              </Button>
              {resultMessage && (
                <styled.span fontSize="sm" color="fg.muted">
                  {resultMessage}
                </styled.span>
              )}
            </HStack>
          )}

          {isLoading ? (
            <VStack
              gap="3"
              width="full"
              p="12"
              alignItems="center"
              justifyContent="center"
              style={{
                border: "1px dashed var(--colors-border-default)",
                borderRadius: "0.75rem",
                backgroundColor: "var(--colors-bg-muted)",
                minHeight: "200px",
              }}
            >
              <Spinner size="lg" />
              <styled.p fontSize="sm" color="fg.muted">
                Loading threads…
              </styled.p>
            </VStack>
          ) : threads.length > 0 ? (
            <PendingReplyThreadList threads={threads} />
          ) : (
            <VStack
              gap="2"
              width="full"
              p="12"
              alignItems="center"
              justifyContent="center"
              style={{
                border: "1px dashed var(--colors-border-default)",
                borderRadius: "0.75rem",
                backgroundColor: "var(--colors-bg-muted)",
                minHeight: "200px",
              }}
            >
              <styled.p fontSize="sm" color="fg.muted">
                No threads pending reply for the selected filters.
              </styled.p>
            </VStack>
          )}
        </>
      )}
    </VStack>
  );
}
