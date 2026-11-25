"use client";

import { useState } from "react";
import { CategoryListOKResponse } from "@/api/openapi-schema";
import { css } from "@/styled-system/css";
import { LStack, HStack } from "@/styled-system/jsx";
import { ChevronDownIcon } from "@/components/ui/icons/Chevron";
import { DiscussionIcon } from "@/components/ui/icons/Discussion";
import { CategoryListTree } from "@/components/category/CategoryList/CategoryList";

type Props = {
  categories: CategoryListOKResponse;
  channelID: string;
  currentCategorySlug?: string;
};

export function TopicsSection({ categories, channelID, currentCategorySlug }: Props) {
  const [isExpanded, setIsExpanded] = useState(true);

  if (!categories?.categories || categories.categories.length === 0) {
    return null;
  }

  return (
    <LStack gap="1">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className={css({
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "1",
          py: "1",
          pr: "1",
          h: "8",
          fontSize: "xs",
          fontWeight: "semibold",
          color: "fg.muted",
          textTransform: "uppercase",
          letterSpacing: "wider",
          background: "transparent",
          border: "none",
          cursor: "pointer",
          w: "full",
          textAlign: "left",
          _hover: {
            color: "fg.subtle",
          },
        })}
      >
        <HStack gap="1">
          <DiscussionIcon width="4" height="4" />
          <span>Topics</span>
        </HStack>
        <ChevronDownIcon
          width="4"
          height="4"
          style={{
            transition: "transform 200ms",
            transform: isExpanded ? "rotate(0deg)" : "rotate(-90deg)",
          }}
        />
      </button>

      {isExpanded && (
        <div
          className={css({
            w: "full",
          })}
        >
          <CategoryListTree
            categories={categories.categories}
            currentCategorySlug={currentCategorySlug}
            channelID={channelID}
            mutate={async () => undefined}
            hideHeader={true}
          />
        </div>
      )}
    </LStack>
  );
}
