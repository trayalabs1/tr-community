"use client";

import { useState } from "react";
import { ChevronDown } from "lucide-react";
import { Account } from "@/api/openapi-schema";
import { HStack, LStack, styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { BookmarkIcon } from "@/components/ui/icons/Bookmark";
import { NotificationIcon } from "@/components/ui/icons/Notification";

interface JourneyStage {
  id: string;
  name: string;
  memberCount: number;
}

interface Topic {
  id: string;
  name: string;
  icon: React.ReactNode;
  memberCount: number;
}

interface MemberSidebarProps {
  account: Account;
  journeyStage?: JourneyStage;
  topics?: Topic[];
}

export function MemberSidebar({
  account,
  journeyStage,
  topics = [],
}: MemberSidebarProps) {
  const [isJourneyOpen, setIsJourneyOpen] = useState(true);
  const [isTopicsOpen, setIsTopicsOpen] = useState(true);

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase();
  };

  return (
    <LStack
      w="full"
      gap="4"
      p="4"
      fontFamily="nunito"
      style={{
        background: "white",
        borderRight: "1px solid #E5E7EB",
      }}
    >
      {/* Member Header */}
      <HStack alignItems="start" gap="3" w="full">
        {/* Avatar */}
        <styled.div
          w="12"
          h="12"
          rounded="full"
          display="flex"
          alignItems="center"
          justifyContent="center"
          flexShrink="0"
          style={{
            background: TRAYA_COLORS.gradient,
          }}
        >
          <styled.span
            fontSize="lg"
            fontWeight="bold"
            color="white"
          >
            {getInitials(account.name || account.handle)}
          </styled.span>
        </styled.div>

        {/* User Info & Icons */}
        <LStack gap="2" w="full" alignItems="start">
          <LStack gap="0.5" w="full">
            <styled.h3
              fontSize="md"
              fontWeight="bold"
              color="fg.default"
              m="0"
            >
              {account.name || account.handle}
            </styled.h3>
            <styled.p
              fontSize="xs"
              color="fg.muted"
              m="0"
              style={{ lineHeight: "1.3", color: TRAYA_COLORS.neutral.text }}
            >
              {account.roles?.[0]?.name || "Member"} • 42 days active
            </styled.p>
          </LStack>

          <HStack gap="2" alignItems="center">
            <styled.button
              display="flex"
              alignItems="center"
              justifyContent="center"
              w="6"
              h="6"
              rounded="md"
              style={{
                background: TRAYA_COLORS.secondary,
                border: "none",
                cursor: "pointer",
                transition: "all 0.2s",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  TRAYA_COLORS.tertiary;
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  TRAYA_COLORS.secondary;
              }}
            >
              <BookmarkIcon width="4" height="4" style={{ color: TRAYA_COLORS.primary }} />
            </styled.button>

            <styled.button
              display="flex"
              alignItems="center"
              justifyContent="center"
              w="6"
              h="6"
              rounded="md"
              style={{
                background: TRAYA_COLORS.secondary,
                border: "none",
                cursor: "pointer",
                transition: "all 0.2s",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  TRAYA_COLORS.tertiary;
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  TRAYA_COLORS.secondary;
              }}
            >
              <NotificationIcon width="4" height="4" style={{ color: TRAYA_COLORS.primary }} />
            </styled.button>
          </HStack>
        </LStack>
      </HStack>

      {/* Divider */}
      <styled.div
        h="0.5"
        w="full"
        style={{ background: TRAYA_COLORS.neutral.light }}
      />

      {/* Journey Stage Section */}
      {journeyStage && (
        <LStack gap="3" w="full">
          <styled.button
            onClick={() => setIsJourneyOpen(!isJourneyOpen)}
            display="flex"
            alignItems="center"
            justifyContent="space-between"
            w="full"
            p="0"
            style={{
              background: "transparent",
              border: "none",
              cursor: "pointer",
              transition: "all 0.2s",
            }}
            onMouseEnter={(e) => {
              const label = (e.currentTarget as HTMLElement).querySelector(
                "span"
              ) as HTMLElement;
              if (label) label.style.color = TRAYA_COLORS.primary;
            }}
            onMouseLeave={(e) => {
              const label = (e.currentTarget as HTMLElement).querySelector(
                "span"
              ) as HTMLElement;
              if (label) label.style.color = TRAYA_COLORS.neutral.textMuted;
            }}
          >
            <styled.span
              fontSize="xs"
              fontWeight="bold"
              textTransform="uppercase"
              style={{
                color: TRAYA_COLORS.neutral.textMuted,
                transition: "all 0.2s",
                letterSpacing: "0.05em",
              }}
            >
              Your Journey Stage
            </styled.span>
            <ChevronDown
              size={16}
              style={{
                color: TRAYA_COLORS.primary,
                transform: isJourneyOpen ? "rotate(0deg)" : "rotate(-90deg)",
                transition: "transform 0.2s",
              }}
            />
          </styled.button>

          {isJourneyOpen && (
            <styled.div
              p="3"
              rounded="lg"
              style={{
                background: TRAYA_COLORS.tertiary,
                border: `1px solid ${TRAYA_COLORS.secondary}`,
                cursor: "pointer",
                transition: "all 0.2s",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  TRAYA_COLORS.secondary;
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background =
                  TRAYA_COLORS.tertiary;
              }}
            >
              <LStack gap="2" alignItems="start">
                <styled.div
                  display="flex"
                  alignItems="center"
                  gap="2"
                  w="full"
                >
                  <styled.div
                    w="8"
                    h="8"
                    rounded="lg"
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    style={{
                      background: TRAYA_COLORS.secondary,
                    }}
                  >
                    <styled.span fontSize="sm">🌱</styled.span>
                  </styled.div>
                  <LStack gap="0" alignItems="start" w="full">
                    <styled.h4
                      fontSize="sm"
                      fontWeight="bold"
                      color="fg.default"
                      m="0"
                    >
                      {journeyStage.name}
                    </styled.h4>
                    <styled.p
                      fontSize="xs"
                      color="fg.muted"
                      m="0"
                      style={{ color: TRAYA_COLORS.neutral.text }}
                    >
                      👥 {journeyStage.memberCount} members
                    </styled.p>
                  </LStack>
                </styled.div>
              </LStack>
            </styled.div>
          )}
        </LStack>
      )}

{/* Topics Section - Hidden for now */}
      {false && topics.length > 0 && (
        <LStack gap="3" w="full">
          <styled.button
            onClick={() => setIsTopicsOpen(!isTopicsOpen)}
            display="flex"
            alignItems="center"
            justifyContent="space-between"
            w="full"
            p="0"
            style={{
              background: "transparent",
              border: "none",
              cursor: "pointer",
              transition: "all 0.2s",
            }}
            onMouseEnter={(e) => {
              const label = (e.currentTarget as HTMLElement).querySelector(
                "span"
              ) as HTMLElement;
              if (label) label.style.color = TRAYA_COLORS.primary;
            }}
            onMouseLeave={(e) => {
              const label = (e.currentTarget as HTMLElement).querySelector(
                "span"
              ) as HTMLElement;
              if (label) label.style.color = TRAYA_COLORS.neutral.textMuted;
            }}
          >
            <styled.span
              fontSize="xs"
              fontWeight="bold"
              textTransform="uppercase"
              style={{
                color: TRAYA_COLORS.neutral.textMuted,
                transition: "all 0.2s",
                letterSpacing: "0.05em",
              }}
            >
              Topics
            </styled.span>
            <ChevronDown
              size={16}
              style={{
                color: TRAYA_COLORS.primary,
                transform: isTopicsOpen ? "rotate(0deg)" : "rotate(-90deg)",
                transition: "transform 0.2s",
              }}
            />
          </styled.button>

          {isTopicsOpen && (
            <LStack gap="2" w="full">
              {topics.map((topic) => (
                <styled.div
                  key={topic.id}
                  p="3"
                  rounded="lg"
                  display="flex"
                  alignItems="center"
                  gap="3"
                  w="full"
                  style={{
                    background: "white",
                    border: `1px solid ${TRAYA_COLORS.secondary}`,
                    cursor: "pointer",
                    transition: "all 0.2s",
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background =
                      TRAYA_COLORS.tertiary;
                    (e.currentTarget as HTMLElement).style.borderColor =
                      TRAYA_COLORS.primary;
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = "white";
                    (e.currentTarget as HTMLElement).style.borderColor =
                      TRAYA_COLORS.secondary;
                  }}
                >
                  {/* Icon */}
                  <styled.div
                    w="8"
                    h="8"
                    rounded="lg"
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    flexShrink="0"
                    style={{
                      background: TRAYA_COLORS.secondary,
                    }}
                  >
                    {topic.icon}
                  </styled.div>

                  {/* Content */}
                  <LStack gap="0" alignItems="start" w="full">
                    <styled.h5
                      fontSize="sm"
                      fontWeight="bold"
                      color="fg.default"
                      m="0"
                    >
                      {topic.name}
                    </styled.h5>
                    <styled.p
                      fontSize="xs"
                      color="fg.muted"
                      m="0"
                      style={{ color: TRAYA_COLORS.neutral.text }}
                    >
                      👥 {topic.memberCount} members
                    </styled.p>
                  </LStack>

                  {/* Arrow */}
                  <styled.div
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    flexShrink="0"
                    style={{
                      color: TRAYA_COLORS.primary,
                    }}
                  >
                    →
                  </styled.div>
                </styled.div>
              ))}
            </LStack>
          )}
        </LStack>
      )}
    </LStack>
  );
}
