"use client";

import { Check, X } from "lucide-react";

import { HStack, VStack, styled } from "@/styled-system/jsx";
import { PRIMARY_TOPICS } from "@/lib/feed/primaryTopic";

const TOPIC_IMAGE_URLS: Record<string, string> = {
  "RESULTS & PROGRESS": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/results_progress.png?v=1788353080",
  "TIPS & EXPERIENCES": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/tips_experiences.png?v=1788353080",
  "HAIRFALL CONCERNS": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/hairfall_concerns.png?v=1788353080",
  "HOW TO USE": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/how_to_use.png?v=1788353080",
  "DANDRUFF & SCALP": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/dandruff_scalp.png?v=1788353080",
  "HAIR REGROWTH": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/hair_regrowth.png?v=1788353080",
  "PRODUCTS & TREATMENT": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/products_treatment.png?v=1788353080",
  "SIDE EFFECTS": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/side_effects.png?v=1788353080",
  "DIET & LIFESTYLE": "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/diet_lifestyle.png?v=1788353080",
  CHALLENGES: "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/challenges.png?v=1788353080",
};

interface CategoryExploreCarouselProps {
  selectedPrimaryTopic: string | null;
  onPrimaryTopicChange: (topic: string | null) => void;
}

export function CategoryExploreCarousel({
  selectedPrimaryTopic,
  onPrimaryTopicChange,
}: CategoryExploreCarouselProps) {
  const handleSelect = (value: string) => {
    onPrimaryTopicChange(selectedPrimaryTopic === value ? null : value);
  };

  const selectedTopicLabel = PRIMARY_TOPICS.find((t) => t.value === selectedPrimaryTopic)?.label;

  return (
    <VStack
      alignItems="start"
      gap="3"
      width="full"
      backgroundColor="white"
      px={{ base: "4", md: "0" }}
      pt={{ base: "7", md: "0" }}
      pb={{ base: "3", md: "0" }}
    >
      <HStack alignItems="center" gap="2" width="full">
        <styled.h2
          fontWeight="semibold"
          style={{ margin: "0", fontSize: "16px", lineHeight: "20px", color: "#2c2c2a" }}
        >
          Explore by Category
        </styled.h2>

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
              padding: "0.25rem 0.625rem",
              borderRadius: "9999px",
              fontWeight: "500",
              fontSize: "13px",
              whiteSpace: "nowrap",
            }}
          >
            {selectedTopicLabel}
            <X size={12} strokeWidth={2} />
          </styled.button>
        )}
      </HStack>

      <styled.div
        width="full"
        overflowX="auto"
        style={{ scrollbarWidth: "none" }}
        css={{ "&::-webkit-scrollbar": { display: "none" } }}
      >
        <HStack gap="3" alignItems="start">
          {PRIMARY_TOPICS.map((topic) => {
            const isSelected = topic.value === selectedPrimaryTopic;
            const imageURL = TOPIC_IMAGE_URLS[topic.value];

            return (
              <styled.button
                key={topic.value}
                type="button"
                onClick={() => handleSelect(topic.value)}
                aria-pressed={isSelected}
                position="relative"
                flexShrink="0"
                cursor="pointer"
                style={{
                  width: "104px",
                  height: "136px",
                  borderRadius: "12px",
                  overflow: "hidden",
                  border: isSelected ? "2px solid #2c2c2a" : "none",
                  padding: "0",
                  background: `linear-gradient(180deg, rgba(0,0,0,0) 40%, rgba(0,0,0,0.65) 100%), url(${imageURL}) center/cover no-repeat`,
                }}
              >
                {isSelected && (
                  <styled.div
                    position="absolute"
                    top="2"
                    left="2"
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    w="5"
                    h="5"
                    borderRadius="full"
                    style={{ backgroundColor: "#22c55e" }}
                  >
                    <Check size={12} color="white" strokeWidth={3} />
                  </styled.div>
                )}

                <styled.p
                  position="absolute"
                  bottom="2"
                  left="2"
                  right="2"
                  textAlign="left"
                  color="white"
                  fontWeight="semibold"
                  style={{ fontSize: "13px", lineHeight: "16px", margin: "0" }}
                >
                  {topic.label}
                </styled.p>
              </styled.button>
            );
          })}
        </HStack>
      </styled.div>
    </VStack>
  );
}
