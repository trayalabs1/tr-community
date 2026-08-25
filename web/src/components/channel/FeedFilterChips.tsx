"use client";

import Link from "next/link";
import { Send } from "lucide-react";

import { BookmarkIcon } from "@/components/ui/icons/Bookmark";
import { HStack, styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { PRIMARY_TOPICS } from "@/lib/feed/primaryTopic";

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

const topicChipBase = {
  display: "flex",
  alignItems: "center",
  height: "36px",
  padding: "0 16px",
  borderRadius: "9999px",
  fontSize: "14px",
  fontWeight: 500,
  lineHeight: "20px",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
  border: "none",
};

interface FeedFilterChipsProps {
  selectedPrimaryTopic: string | null;
  onPrimaryTopicChange: (topic: string | null) => void;
}

export function FeedFilterChips({ selectedPrimaryTopic, onPrimaryTopicChange }: FeedFilterChipsProps) {
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

      <styled.button
        onClick={() => onPrimaryTopicChange(null)}
        style={{
          ...topicChipBase,
          backgroundColor: selectedPrimaryTopic === null ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
          color: selectedPrimaryTopic === null ? "white" : TRAYA_COLORS.primary,
        }}
      >
        All
      </styled.button>

      {PRIMARY_TOPICS.map((topic) => (
        <styled.button
          key={topic.value}
          onClick={() => onPrimaryTopicChange(topic.value)}
          style={{
            ...topicChipBase,
            backgroundColor: selectedPrimaryTopic === topic.value ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
            color: selectedPrimaryTopic === topic.value ? "white" : TRAYA_COLORS.primary,
          }}
        >
          {topic.label}
        </styled.button>
      ))}
    </HStack>
  );
}
