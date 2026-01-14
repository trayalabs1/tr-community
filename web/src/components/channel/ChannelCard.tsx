"use client";

import { HStack, VStack, styled } from "@/styled-system/jsx";
import { ChevronRightIcon } from "@/components/ui/icons/Chevron";
import { ProfileIcon } from "@/components/ui/icons/Profile";
import { MembersIcon } from "@/components/ui/icons/Members";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { getAssetURL } from "@/utils/asset";

interface ChannelCardProps {
  id: string;
  name: string;
  icon?: {
    path: string;
  };
  memberCount?: number;
  isActive?: boolean;
  isCohort?: boolean;
  onClick?: () => void;
}

export function ChannelCard({
  id,
  name,
  icon,
  memberCount = 0,
  isActive = false,
  isCohort = false,
  onClick,
}: ChannelCardProps) {
  const iconSize = isCohort ? "14" : "12";
  const padding = isCohort ? "4" : "3";

  return (
    <styled.button
      display="flex"
      alignItems="center"
      gap="3"
      rounded="xl"
      width="full"
      style={{
        backgroundColor: isActive ? TRAYA_COLORS.tertiary : "transparent",
        padding: isCohort ? "1rem" : "0.75rem",
        border: "none",
        cursor: "pointer",
        transition: "all 0.2s ease-in-out",
      }}
      onMouseEnter={(e) => {
        if (!isActive) {
          (e.currentTarget as HTMLElement).style.backgroundColor =
            TRAYA_COLORS.tertiary;
        }
      }}
      onMouseLeave={(e) => {
        if (!isActive) {
          (e.currentTarget as HTMLElement).style.backgroundColor =
            "transparent";
        }
      }}
      onClick={onClick}
      aria-label={`${name} channel`}
    >
      {/* Icon Container */}
      <styled.div
        rounded="2xl"
        display="flex"
        alignItems="center"
        justifyContent="center"
        flexShrink="0"
        style={{
          width: isCohort ? "56px" : "48px",
          height: isCohort ? "56px" : "48px",
          background: isActive
            ? `linear-gradient(135deg, ${TRAYA_COLORS.primary}30, ${TRAYA_COLORS.primary}15)`
            : `linear-gradient(135deg, ${TRAYA_COLORS.primary}10, ${TRAYA_COLORS.tertiary})`,
          fontSize: "24px",
          overflow: "hidden",
        }}
      >
        {icon ? (
          <styled.img
            src={getAssetURL(icon.path)}
            alt={name}
            w={isCohort ? "14" : "12"}
            h={isCohort ? "14" : "12"}
            rounded="2xl"
            objectFit="cover"
          />
        ) : (
          <ProfileIcon
            width={isCohort ? "8" : "6"}
            height={isCohort ? "8" : "6"}
            style={{ color: TRAYA_COLORS.primary }}
          />
        )}
      </styled.div>

      {/* Content */}
      <VStack alignItems="start" gap="0.5" flex="1">
        <styled.span
          fontWeight="semibold"
          style={{
            fontSize: isCohort ? "16px" : "14px",
            color: "var(--colors-fg-default)",
            margin: "0",
          }}
        >
          {name}
        </styled.span>
        <HStack gap="1" alignItems="center">
          <MembersIcon
            width="3.5"
            height="3.5"
            style={{ color: "var(--colors-fg-muted)" }}
          />
          <styled.span
            fontSize="xs"
            style={{
              color: "var(--colors-fg-muted)",
              margin: "0",
            }}
          >
            {memberCount.toLocaleString()} members
          </styled.span>
        </HStack>
      </VStack>

      {/* Chevron */}
      <ChevronRightIcon
        width="5"
        height="5"
        style={{ color: "var(--colors-fg-muted)", flexShrink: 0 }}
      />
    </styled.button>
  );
}
