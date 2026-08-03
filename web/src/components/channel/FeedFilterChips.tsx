"use client";

import Link from "next/link";
import { Portal } from "@ark-ui/react";
import { Send } from "lucide-react";

import { Category } from "@/api/openapi-schema";
import { ChevronDownIcon } from "@/components/ui/icons/Chevron";
import { BookmarkIcon } from "@/components/ui/icons/Bookmark";
import * as Menu from "@/components/ui/menu";
import { HStack, styled } from "@/styled-system/jsx";

type Props = {
  categories: Category[];
  selectedCategorySlug: string | null;
  onCategoryChange: (slug: string | null) => void;
};

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

export function FeedFilterChips({
  categories,
  selectedCategorySlug,
  onCategoryChange,
}: Props) {
  const selected = categories.find((c) => c.slug === selectedCategorySlug);
  const label = selected?.name ?? "All";

  return (
    <HStack gap="2" width="full" overflowX="auto">
      <Menu.Root>
        <Menu.Trigger asChild>
          <styled.button
            type="button"
            flexShrink="0"
            style={{
              ...chipBase,
              backgroundColor: "#404040",
              color: "#ffffff",
              border: "none",
            }}
          >
            {label}
            <ChevronDownIcon width="4" height="4" />
          </styled.button>
        </Menu.Trigger>
        <Portal>
          <Menu.Positioner>
            <Menu.Content>
              <Menu.Item
                value="__all"
                onClick={() => onCategoryChange(null)}
              >
                All
              </Menu.Item>
              {categories.map((c) => (
                <Menu.Item
                  key={c.slug}
                  value={c.slug}
                  onClick={() => onCategoryChange(c.slug)}
                >
                  {c.name}
                </Menu.Item>
              ))}
            </Menu.Content>
          </Menu.Positioner>
        </Portal>
      </Menu.Root>

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
    </HStack>
  );
}
