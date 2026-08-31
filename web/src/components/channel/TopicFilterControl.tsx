"use client";

import { useMemo } from "react";
import { createListCollection, type SelectValueChangeDetails } from "@ark-ui/react";
import { X } from "lucide-react";

import { ChevronDownIcon } from "@/components/ui/icons/Chevron";
import * as Select from "@/components/ui/select";
import { styled } from "@/styled-system/jsx";
import { PRIMARY_TOPICS } from "@/lib/feed/primaryTopic";

const ALL_TOPICS_VALUE = "__all";

interface TopicFilterControlProps {
  selectedPrimaryTopic: string | null;
  onPrimaryTopicChange: (topic: string | null) => void;
}

// Matches the Figma reference (community feed filter row): a fixed "All"
// trigger that always reads "All" and opens the topic picker, plus a
// separate dismissible chip showing the active topic once one is selected —
// rather than the trigger's own label changing to the selected value.
export function TopicFilterControl({ selectedPrimaryTopic, onPrimaryTopicChange }: TopicFilterControlProps) {
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

  const selectedTopicLabel = PRIMARY_TOPICS.find((t) => t.value === selectedPrimaryTopic)?.label;

  return (
    <>
      <Select.Root
        size="sm"
        collection={topicCollection}
        value={[selectedPrimaryTopic ?? ALL_TOPICS_VALUE]}
        positioning={{ sameWidth: false }}
        onValueChange={handleTopicChange}
        width="auto"
        gap="0"
        flexShrink="0"
      >
        <Select.Control>
          <Select.Trigger asChild>
            <styled.button
              type="button"
              display="flex"
              alignItems="center"
              justifyContent="center"
              gap="1"
              flexShrink="0"
              cursor="pointer"
              style={{
                backgroundColor: "#2c2c2a",
                color: "white",
                border: "none",
                padding: "0.375rem 0.75rem",
                width: "5rem",
                borderRadius: "9999px",
                fontWeight: "500",
                fontSize: "14px",
              }}
            >
              All
              <ChevronDownIcon width="4" height="4" color="white" />
            </styled.button>
          </Select.Trigger>
        </Select.Control>
        <Select.Positioner>
          <Select.Content
            padding="0"
            borderRadius="l3"
            overflow="hidden"
            style={{ minWidth: "10rem" }}
          >
            {topicCollection.items.map((item, index) => (
              <Select.Item
                key={item.value}
                item={item}
                borderRadius="[0]"
                style={{
                  padding: "0.875rem 1.25rem",
                  fontSize: "16px",
                  color: "#404040",
                  borderBottom:
                    index < topicCollection.items.length - 1 ? "1px solid #ECECEC" : "none",
                }}
                css={{
                  "&[data-state=checked]": { fontWeight: "semibold", color: "[#1a1a1a]" },
                  "&:hover, &[data-highlighted]": { background: "[#F7F7F7]" },
                }}
              >
                <Select.ItemText>{item.label}</Select.ItemText>
              </Select.Item>
            ))}
          </Select.Content>
        </Select.Positioner>
      </Select.Root>

      {selectedTopicLabel && (
        <styled.button
          type="button"
          onClick={() => onPrimaryTopicChange(null)}
          display="flex"
          alignItems="center"
          gap="1"
          flexShrink="0"
          cursor="pointer"
          style={{
            backgroundColor: "#ECECEC",
            color: "#404040",
            border: "none",
            padding: "0.375rem 0.75rem",
            borderRadius: "9999px",
            fontWeight: "500",
            fontSize: "14px",
            whiteSpace: "nowrap",
          }}
        >
          {selectedTopicLabel}
          <X size={14} strokeWidth={2} />
        </styled.button>
      )}
    </>
  );
}
