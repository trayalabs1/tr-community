"use client";

import { useState } from "react";
import { HStack, LStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { ChevronDownIcon } from "@/components/ui/icons/Chevron";
import { InfoIcon } from "@/components/ui/icons/Info";
import { HeaderWithBackArrow } from "@/components/site/Header";
import { FAQ_DATA, FAQ_CATEGORIES } from "@/content/faq";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { Props, useInfoScreen } from "./useInfoScreen";

export function InfoScreen(props: Props) {
  useInfoScreen(props);

  // Track selected category for filtering
  const [selectedCategory, setSelectedCategory] = useState("All");
  // Track which FAQ item is expanded
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Filter FAQ data based on selected category
  const filteredFAQ =
    selectedCategory === "All"
      ? FAQ_DATA
      : FAQ_DATA.filter((item) => item.category === selectedCategory);

  return (
    <LStack gap="0" width="full" height="screen" style={{ backgroundColor: "white" }}>
      <HeaderWithBackArrow
        title="Info"
        subtitle="Frequently asked questions"
        headerIcon={<InfoIcon width="5" height="5" style={{ color: "#ffffff" }} />}
        headerIconBackground={`${TRAYA_COLORS.primary}FF`}
        mobileOnly
        isSticky
      />

      {/* Content Wrapper */}
      <styled.div
        flex="1"
        width="full"
        style={{
          overflowY: "auto",
          minHeight: 0,
        }}
      >
        {/* Desktop Header - Only on Desktop */}
        <styled.div display={{ base: "none", md: "block" }} px="4" py="4">
          <WStack alignItems="start">
            {/* Header with info icon and title */}
            <HStack gap="2" alignItems="center">
              <styled.div
                w="8"
                h="8"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                style={{
                  backgroundColor: `${TRAYA_COLORS.primary}20`,
                }}
              >
                <InfoIcon width="5" height="5" style={{ color: TRAYA_COLORS.primary }} />
              </styled.div>
              <VStack gap="0" alignItems="start">
                <styled.h2 fontSize="lg" fontWeight="semibold" color="fg.default">
                  Info
                </styled.h2>
                <styled.p fontSize="sm" color="fg.muted">
                  Frequently asked questions
                </styled.p>
              </VStack>
            </HStack>
          </WStack>
        </styled.div>

        {/* Info Content - FAQ Library */}
        <VStack gap="4" width="full" px="4" py="4">

          {/* Category filter buttons - horizontally scrollable */}
          <HStack
            gap="2"
            width="full"
            style={{
              overflowX: "auto",
              paddingBottom: "0.5rem",
              scrollBehavior: "smooth",
            }}
          >
            {FAQ_CATEGORIES.map((category) => (
              <styled.button
                key={category}
                px="3"
                py="2"
                rounded="full"
                fontSize="sm"
                fontWeight="medium"
                onClick={() => setSelectedCategory(category)}
                style={{
                  backgroundColor:
                    selectedCategory === category
                      ? TRAYA_COLORS.primary
                      : "rgba(0, 0, 0, 0.05)",
                  color:
                    selectedCategory === category
                      ? "white"
                      : "var(--colors-fg-default)",
                  border: "none",
                  cursor: "pointer",
                  transition: "all 0.2s",
                  whiteSpace: "nowrap",
                }}
                onMouseEnter={(e) => {
                  if (selectedCategory !== category) {
                    (e.currentTarget as HTMLElement).style.backgroundColor =
                      "rgba(0, 0, 0, 0.08)";
                  }
                }}
                onMouseLeave={(e) => {
                  if (selectedCategory !== category) {
                    (e.currentTarget as HTMLElement).style.backgroundColor =
                      "rgba(0, 0, 0, 0.05)";
                  }
                }}
              >
                {category}
              </styled.button>
            ))}
          </HStack>

          {/* FAQ items list */}
          <VStack gap="3" width="full">
            {filteredFAQ.map((item) => (
              <styled.div
                key={item.id}
                width="full"
                rounded="lg"
                overflow="hidden"
                style={{
                  backgroundColor: "rgba(0, 0, 0, 0.02)",
                  border: "1px solid rgba(0, 0, 0, 0.05)",
                  transition: "all 0.2s",
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.backgroundColor =
                    "rgba(0, 0, 0, 0.04)";
                  (e.currentTarget as HTMLElement).style.borderColor =
                    "rgba(0, 0, 0, 0.1)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.backgroundColor =
                    "rgba(0, 0, 0, 0.02)";
                  (e.currentTarget as HTMLElement).style.borderColor =
                    "rgba(0, 0, 0, 0.05)";
                }}
              >
                {/* FAQ question button - toggles expansion */}
                <styled.button
                  width="full"
                  px="4"
                  py="4"
                  display="flex"
                  alignItems="center"
                  justifyContent="space-between"
                  gap="2"
                  onClick={() =>
                    setExpandedId(expandedId === item.id ? null : item.id)
                  }
                  style={{
                    background: "transparent",
                    border: "none",
                    cursor: "pointer",
                    padding: "1rem",
                  }}
                >
                  <styled.span
                    fontSize="sm"
                    fontWeight="medium"
                    color="fg.default"
                    textAlign="left"
                  >
                    {item.question}
                  </styled.span>
                  {/* Chevron icon with rotation animation */}
                  <styled.span
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    style={{
                      flexShrink: 0,
                      transform:
                        expandedId === item.id ? "rotate(180deg)" : "rotate(0deg)",
                      transition: "transform 0.2s",
                    }}
                  >
                    <ChevronDownIcon width="5" height="5" />
                  </styled.span>
                </styled.button>

                {/* FAQ answer - shown when expanded */}
                {expandedId === item.id && (
                  <styled.div
                    px="4"
                    pb="4"
                    fontSize="sm"
                    color="fg.muted"
                    style={{
                      borderTopWidth: "1px",
                      borderTopColor: "rgba(0, 0, 0, 0.05)",
                      paddingTop: "1rem",
                      lineHeight: "1.6",
                    }}
                  >
                    {item.answer}
                  </styled.div>
                )}
              </styled.div>
            ))}
          </VStack>
        </VStack>
      </styled.div>
    </LStack>
  );
}
