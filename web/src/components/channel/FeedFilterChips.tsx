"use client";

import Link from "next/link";
import { useMemo } from "react";
import { createListCollection, type SelectValueChangeDetails } from "@ark-ui/react";
import { Send } from "lucide-react";

import { BookmarkIcon } from "@/components/ui/icons/Bookmark";
import { CheckIcon } from "@/components/ui/icons/Check";
import { SelectIcon } from "@/components/ui/icons/Select";
import * as Select from "@/components/ui/select";
import { HStack, styled } from "@/styled-system/jsx";
import { PRIMARY_TOPICS } from "@/lib/feed/primaryTopic";

const ALL_TOPICS_VALUE = "__all";

const chipBase = {
  display: "flex",
  alignItems: "center",
  gap: "6px",
  height: "36px",
  padding: "0 16px",
  borderRadius: "9999px",
  fontSize: "14px",
  fontWeight: 600,
  lineHeight: "20px",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
};

interface FeedFilterChipsProps {
  selectedPrimaryTopic: string | null;
  onPrimaryTopicChange: (topic: string | null) => void;
}

export function FeedFilterChips({ selectedPrimaryTopic, onPrimaryTopicChange }: FeedFilterChipsProps) {
  const topicCollection = useMemo(
    () =>
      createListCollection({
        items: [
          { label: "All", value: ALL_TOPICS_VALUE },
          ...PRIMARY_TOPICS,
        ],
      }),
    [],
  );

  const handleTopicChange = ({ value }: SelectValueChangeDetails) => {
    const [selected] = value;
    if (!selected || selected === ALL_TOPICS_VALUE) {
      onPrimaryTopicChange(null);
      return;
    }
    onPrimaryTopicChange(selected);
  };

  return (
    <HStack gap="2" width="full" flexWrap="wrap">
      <Link href="/my-posts" style={{ textDecoration: "none", flexShrink: 0 }}>
        <styled.span
          style={{
            ...chipBase,
            backgroundColor: "white",
            color: "#404040",
            border: "1px solid #dedede",
          }}
        >
          <Send size={16} />
          My posts
        </styled.span>
      </Link>

      <Link href="/c" style={{ textDecoration: "none", flexShrink: 0 }}>
        <styled.span
          style={{
            ...chipBase,
            backgroundColor: "white",
            color: "#404040",
            border: "1px solid #dedede",
          }}
        >
          <BookmarkIcon width="4" height="4" />
          My Bookmarks
        </styled.span>
      </Link>

      <Select.Root
        size="sm"
        collection={topicCollection}
        value={[selectedPrimaryTopic ?? ALL_TOPICS_VALUE]}
        positioning={{ sameWidth: false }}
        onValueChange={handleTopicChange}
      >
        <Select.Control>
          <Select.Trigger>
            <Select.ValueText placeholder="All" />
            <SelectIcon />
          </Select.Trigger>
        </Select.Control>
        <Select.Positioner>
          <Select.Content>
            {topicCollection.items.map((item) => (
              <Select.Item key={item.value} item={item}>
                <Select.ItemText>{item.label}</Select.ItemText>
                <Select.ItemIndicator>
                  <CheckIcon />
                </Select.ItemIndicator>
              </Select.Item>
            ))}
          </Select.Content>
        </Select.Positioner>
      </Select.Root>
    </HStack>
  );
}
