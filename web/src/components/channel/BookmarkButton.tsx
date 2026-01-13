"use client";

import Link from "next/link";
import { BookmarkIcon } from "@/components/ui/icons/Bookmark";
import { styled } from "@/styled-system/jsx";

export function BookmarkButton() {
  return (
    <Link href="/c">
      <styled.button
        display="flex"
        alignItems="center"
        justifyContent="center"
        w="10"
        h="10"
        rounded="md"
        style={{
          background: "transparent",
          border: "none",
          cursor: "pointer",
          color: "var(--colors-fg-default)",
          padding: "0",
          transition: "color 0.2s ease-in-out",
        }}
        onMouseEnter={(e) => {
          (e.currentTarget as HTMLElement).style.color = "#4a9d6f";
        }}
        onMouseLeave={(e) => {
          (e.currentTarget as HTMLElement).style.color = "var(--colors-fg-default)";
        }}
        aria-label="Collections"
      >
        <BookmarkIcon width="5" height="5" />
      </styled.button>
    </Link>
  );
}
