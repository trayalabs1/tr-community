"use client";

import { SelectValueChangeDetails, createListCollection } from "@ark-ui/react";
import { today, getLocalTimeZone, type DateValue } from "@internationalized/date";
import Link from "next/link";
import { useCallback, useMemo, useState } from "react";

import { useAdminAccountList, useAdminReplyList } from "@/api/openapi-client/admin";
import { EmptyState } from "@/components/site/EmptyState";
import { LoadingBanner } from "@/components/site/Loading";
import { Timestamp } from "@/components/site/Timestamp";
import { DateRangePicker } from "@/components/ui/date-picker";
import { SelectIcon } from "@/components/ui/icons/Select";
import * as Select from "@/components/ui/select";
import { HStack, styled, VStack } from "@/styled-system/jsx";

const ALL_ADMINS = "__all";

// This is a retrospective activity view rather than a live work queue, so it
// allows a much wider lookback than the pending-reply tab's three days.
const MAX_LOOKBACK_DAYS = 90;

function startOfDayISO(value: DateValue) {
  const d = value.toDate(getLocalTimeZone());
  d.setHours(0, 0, 0, 0);
  return d.toISOString();
}

function endOfDayISO(value: DateValue) {
  const d = value.toDate(getLocalTimeZone());
  d.setHours(23, 59, 59, 999);
  return d.toISOString();
}

export function AdminRepliesPanel() {
  const todayVal = useMemo(() => today(getLocalTimeZone()), []);

  const [selectedAdmin, setSelectedAdmin] = useState<string | null>(null);
  const [range, setRange] = useState<{ createdAfter?: string; createdBefore?: string }>({});

  const { data: adminData } = useAdminAccountList();

  const adminCollection = useMemo(() => {
    const admins = adminData?.admins ?? [];
    return createListCollection({
      items: [
        { value: ALL_ADMINS, label: "All admins" },
        ...admins.map((a) => ({
          value: a.id,
          label: a.name ? `${a.name} (@${a.handle})` : `@${a.handle}`,
        })),
      ],
    });
  }, [adminData]);

  const handleAdminChange = useCallback(({ value }: SelectValueChangeDetails) => {
    const [selected] = value;
    setSelectedAdmin(!selected || selected === ALL_ADMINS ? null : selected);
  }, []);

  const handleDateChange = useCallback(({ value }: { value: DateValue[] }) => {
    const [start, end] = value;

    if (!start) {
      setRange({});
      return;
    }

    if (!end) return;

    const [earlier, later] = start.compare(end) <= 0 ? [start, end] : [end, start];

    setRange({
      createdAfter: startOfDayISO(earlier),
      createdBefore: endOfDayISO(later),
    });
  }, []);

  const { data, isValidating } = useAdminReplyList({
    ...(selectedAdmin && { replied_by: selectedAdmin }),
    ...(range.createdAfter && { created_after: range.createdAfter }),
    ...(range.createdBefore && { created_before: range.createdBefore }),
  });

  const replies = data?.replies ?? [];
  const selectValue = selectedAdmin ?? ALL_ADMINS;

  return (
    <VStack gap="4" width="full" alignItems="start">
      <HStack gap="3" width="full" flexWrap="wrap" alignItems="end">
        <VStack alignItems="start" gap="2">
          <styled.label
            fontSize="xs"
            fontWeight="semibold"
            color="fg.muted"
            textTransform="uppercase"
          >
            Replied By
          </styled.label>
          <Select.Root
            size="sm"
            collection={adminCollection}
            value={[selectValue]}
            onValueChange={handleAdminChange}
            width="64"
          >
            <Select.Control>
              <Select.Trigger>
                <Select.ValueText placeholder="All admins" />
                <SelectIcon />
              </Select.Trigger>
            </Select.Control>
            <Select.Positioner>
              <Select.Content>
                {adminCollection.items.map((item) => (
                  <Select.Item key={item.value} item={item}>
                    <Select.ItemText>{item.label}</Select.ItemText>
                  </Select.Item>
                ))}
              </Select.Content>
            </Select.Positioner>
          </Select.Root>
        </VStack>

        <VStack alignItems="start" gap="2">
          <styled.label
            fontSize="xs"
            fontWeight="semibold"
            color="fg.muted"
            textTransform="uppercase"
          >
            Reply Date
          </styled.label>
          <DateRangePicker
            hideInputs={true}
            min={todayVal.subtract({ days: MAX_LOOKBACK_DAYS })}
            max={todayVal}
            onValueChange={handleDateChange}
          />
        </VStack>
      </HStack>

      {isValidating && replies.length === 0 ? (
        <LoadingBanner />
      ) : replies.length === 0 ? (
        <EmptyState hideContributionLabel>
          No admin replies match these filters.
        </EmptyState>
      ) : (
        <VStack gap="3" width="full">
          {replies.map((reply) => (
            <VStack
              key={reply.id}
              alignItems="start"
              gap="1"
              width="full"
              p="4"
              style={{
                border: "1px solid var(--colors-border-default)",
                borderRadius: "0.5rem",
                backgroundColor: "var(--colors-bg-default)",
              }}
            >
              <styled.p
                fontSize="sm"
                color="fg.default"
                style={{ wordBreak: "break-word" }}
              >
                {reply.description || "(no content)"}
              </styled.p>

              <HStack gap="2" flexWrap="wrap" fontSize="xs" color="fg.muted">
                <styled.span>
                  on{" "}
                  <Link href={`/t/${reply.root_slug}`}>
                    <styled.span textDecoration="underline">
                      {reply.title || "Untitled post"}
                    </styled.span>
                  </Link>
                </styled.span>
                <styled.span>·</styled.span>
                <styled.span>@{reply.author.handle}</styled.span>
                <styled.span>·</styled.span>
                <Timestamp created={reply.createdAt} />
              </HStack>
            </VStack>
          ))}
        </VStack>
      )}
    </VStack>
  );
}
