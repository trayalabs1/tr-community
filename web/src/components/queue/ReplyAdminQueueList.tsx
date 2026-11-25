"use client";

import Link from "next/link";
import { useState } from "react";
import { styled, VStack, HStack } from "@/styled-system/jsx";
import { ReplyQueueEntry } from "@/api/openapi-schema";
import { EmptyState } from "@/components/site/EmptyState";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

type Props = {
  entries: ReplyQueueEntry[];
  channelMap: Map<string, { name?: string }>;
  onDismiss: (id: string) => Promise<void>;
};

export function ReplyAdminQueueList({ entries, channelMap, onDismiss }: Props) {
  const [dismissing, setDismissing] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);
  const { trackViewPostClicked, trackDismissClicked } = useEventTracking();

  if (entries.length === 0) {
    return (
      <EmptyState hideContributionLabel>
        No replies pending admin attention.
      </EmptyState>
    );
  }

  return (
    <VStack gap="3" width="full">
      {error && (
        <styled.p fontSize="sm" color="fg.error">
          {error}
        </styled.p>
      )}
      {entries.map((entry) => {
        const channel = channelMap.get(entry.channel_id);
        const isDismissing = dismissing.has(entry.id);

        return (
          <HStack
            key={entry.id}
            width="full"
            p="4"
            gap="4"
            justifyContent="space-between"
            alignItems="start"
            style={{
              border: "1px solid var(--colors-border-default)",
              borderRadius: "0.5rem",
              backgroundColor: "var(--colors-bg-default)",
            }}
          >
            <VStack alignItems="start" gap="1" flex="1" minWidth="0">
              <styled.p
                fontSize="sm"
                color="fg.default"
                style={{ wordBreak: "break-word" }}
              >
                {entry.content_snippet || "(no content)"}
              </styled.p>
              <HStack gap="3" fontSize="xs" color="fg.muted">
                <Link
                  href={`/channels/${entry.channel_id}/threads/locate/${entry.reply_id}`}
                  style={{ textDecoration: "underline" }}
                  onClick={() => trackViewPostClicked(entry.reply_id, entry.channel_id)}
                >
                  View reply
                </Link>
                {channel?.name && (
                  <styled.span>{channel.name}</styled.span>
                )}
                <styled.span>
                  {new Date(entry.created_at).toLocaleDateString()}
                </styled.span>
              </HStack>
            </VStack>
            <styled.button
              onClick={async () => {
                trackDismissClicked(entry.reply_id, entry.channel_id);
                setDismissing((prev) => new Set(prev).add(entry.id));
                setError(null);
                try {
                  await onDismiss(entry.id);
                } catch {
                  setDismissing((prev) => {
                    const next = new Set(prev);
                    next.delete(entry.id);
                    return next;
                  });
                  setError("Failed to dismiss. Please try again.");
                }
              }}
              disabled={isDismissing}
              fontSize="xs"
              fontWeight="semibold"
              cursor={isDismissing ? "not-allowed" : "pointer"}
              px="3"
              py="1.5"
              rounded="md"
              flexShrink="0"
              style={{
                backgroundColor: "var(--colors-bg-muted)",
                border: "1px solid var(--colors-border-default)",
                color: isDismissing
                  ? "var(--colors-fg-muted)"
                  : "var(--colors-fg-default)",
              }}
            >
              {isDismissing ? "Dismissing…" : "Dismiss"}
            </styled.button>
          </HStack>
        );
      })}
    </VStack>
  );
}
