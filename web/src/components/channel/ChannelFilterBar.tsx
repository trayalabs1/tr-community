"use client";

import { useState } from "react";
import { Filter, X } from "lucide-react";

import { Category, Permission } from "@/api/openapi-schema";
import { HStack, VStack, styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { hasPermission } from "@/utils/permissions";
import { useSession } from "@/auth";
import { ThreadCreateTrigger } from "@/components/thread/ThreadCreate/ThreadCreateTrigger";

interface ChannelFilterBarProps {
  channelID: string;
  categories: Category[] | undefined;
  selectedCategorySlug: string | null;
  selectedVisibility: string | null;
  onCategoryChange: (slug: string | null) => void;
  onVisibilityChange: (visibility: string | null) => void;
}

export function ChannelFilterBar({
  channelID,
  categories,
  selectedCategorySlug,
  selectedVisibility,
  onCategoryChange,
  onVisibilityChange,
}: ChannelFilterBarProps) {
  const [showFilters, setShowFilters] = useState(false);
  const session = useSession();
  const canManagePosts = hasPermission(session, Permission.MANAGE_POSTS);

  const hasActiveFilters = selectedCategorySlug || selectedVisibility;
  const activeFilterCount = [selectedCategorySlug, selectedVisibility].filter(Boolean).length;

  const clearFilters = () => {
    onCategoryChange(null);
    onVisibilityChange(null);
  };

  return (
    <VStack alignItems="start" gap="3" width="full">
      {/* Filter and Create Post Button Row */}
      <HStack alignItems="center" gap="2" width="full">
        <styled.button
          onClick={() => setShowFilters(!showFilters)}
          display="flex"
          alignItems="center"
          gap="2"
          fontSize="sm"
          fontWeight="medium"
          cursor="pointer"
          style={{
            backgroundColor: "#ecf4eecc",
            color:  "#2d5340",
            border: "none",
            padding: "0.5rem 0.75rem",
            borderRadius: "0.875rem",
            transition: "all 0.2s ease-in-out",
          }}
        >
          <Filter size={16} strokeWidth={2}/>
          <styled.span>Filters</styled.span>
          {hasActiveFilters && (
            <styled.div
              display="flex"
              alignItems="center"
              justifyContent="center"
              style={{
                width: "1.25rem",
                height: "1.25rem",
                borderRadius: "50%",
                backgroundColor: "rgba(255, 255, 255, 0.2)",
                fontSize: "11px",
                fontWeight: "600",
                color: "white",
              }}
            >
              {activeFilterCount}
            </styled.div>
          )}
        </styled.button>

        {hasActiveFilters && (
          <styled.button
            onClick={clearFilters}
            display="flex"
            alignItems="center"
            gap="1"
            fontSize="sm"
            cursor="pointer"
            style={{
              backgroundColor: "transparent",
              color: "var(--colors-fg-muted)",
              border: "none",
              padding: "0.5rem 0.5rem",
              borderRadius: "0.875rem",
              transition: "all 0.2s ease-in-out",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.color = "var(--colors-fg-default)";
              e.currentTarget.style.backgroundColor = TRAYA_COLORS.tertiary;
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.color = "var(--colors-fg-muted)";
              e.currentTarget.style.backgroundColor = "transparent";
            }}
          >
            <X size={16} strokeWidth={2} />
            <styled.span>Clear</styled.span>
          </styled.button>
        )}

      </HStack>

      {/* Divider - extends beyond parent padding */}
      <styled.div
        style={{
          height: "1px",
          backgroundColor: "#e5e5e5",
          marginLeft: "-1rem",
          marginRight: "-1rem",
          width: "calc(100% + 2rem)",
        }}
      />

      {/* Create Post CTA */}
      <ThreadCreateTrigger channelID={channelID} />

      {/* Expanded Filter View */}
      {showFilters && (
        <VStack
          alignItems="start"
          gap="3"
          width="full"
          style={{
            backgroundColor: "var(--colors-bg-default)",
            padding: "1rem",
            borderRadius: "0.875rem",
            border: `1px solid var(--colors-border-default)`,
            animation: "slideDown 0.2s ease-out",
          }}
        >
          {/* Topic Section */}
          {categories && categories.length > 0 && (
            <VStack alignItems="start" gap="2" width="full">
              <styled.label
                fontSize="xs"
                fontWeight="semibold"
                color="fg.muted"
                textTransform="uppercase"
                style={{ letterSpacing: "0.05em" }}
              >
                Topic
              </styled.label>
              <HStack gap="2" flexWrap="wrap">
                <styled.button
                  onClick={() => onCategoryChange(null)}
                  fontSize="sm"
                  cursor="pointer"
                  transition="all"
                  style={{
                    backgroundColor: selectedCategorySlug === null ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
                    color: selectedCategorySlug === null ? "white" : TRAYA_COLORS.primary,
                    border: "none",
                    padding: "0.375rem 0.75rem",
                    borderRadius: "9999px",
                    fontWeight: "500",
                    fontSize: "14px",
                  }}
                >
                  All
                </styled.button>
                {categories.map((category) => (
                  <styled.button
                    key={category.id}
                    onClick={() => onCategoryChange(category.slug)}
                    fontSize="sm"
                    cursor="pointer"
                    transition="all"
                    style={{
                      backgroundColor: selectedCategorySlug === category.slug ? TRAYA_COLORS.primary : `${TRAYA_COLORS.tertiary}`,
                      color: selectedCategorySlug === category.slug ? "white" : TRAYA_COLORS.primary,
                      border: "none",
                      padding: "0.375rem 0.75rem",
                      borderRadius: "9999px",
                      fontWeight: "500",
                      fontSize: "14px",
                    }}
                  >
                    {category.name}
                  </styled.button>
                ))}
              </HStack>
            </VStack>
          )}

          {/* Status Section */}
          {canManagePosts && (
            <VStack alignItems="start" gap="2" width="full">
              <styled.label
                fontSize="xs"
                fontWeight="semibold"
                color="fg.muted"
                textTransform="uppercase"
                style={{ letterSpacing: "0.05em" }}
              >
                Status
              </styled.label>
              <HStack gap="2" flexWrap="wrap">
                <styled.button
                  onClick={() => onVisibilityChange(null)}
                  fontSize="sm"
                  cursor="pointer"
                  transition="all"
                  style={{
                    backgroundColor: selectedVisibility === null ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
                    color: selectedVisibility === null ? "white" : TRAYA_COLORS.primary,
                    border: "none",
                    padding: "0.375rem 0.75rem",
                    borderRadius: "9999px",
                    fontWeight: "500",
                    fontSize: "14px",
                  }}
                >
                  All
                </styled.button>
                <styled.button
                  onClick={() => onVisibilityChange("review")}
                  fontSize="sm"
                  cursor="pointer"
                  transition="all"
                  style={{
                    backgroundColor: selectedVisibility === "review" ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
                    color: selectedVisibility === "review" ? "white" : TRAYA_COLORS.primary,
                    border: "none",
                    padding: "0.375rem 0.75rem",
                    borderRadius: "9999px",
                    fontWeight: "500",
                    fontSize: "14px",
                  }}
                >
                  In Review
                </styled.button>
              </HStack>
            </VStack>
          )}
        </VStack>
      )}
    </VStack>
  );
}
