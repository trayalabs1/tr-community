"use client";

import Link from "next/link";
import { useLayoutEffect, useMemo, useRef, useState } from "react";

import { styled } from "@/styled-system/jsx";

type Props = {
  /** Full post text. If it contains HTML it is stripped to plain text. */
  text: string;
  /** Used when `text` strips down to empty (e.g. an image-only HTML body). */
  fallback?: string;
  /** Lines to show when collapsed. */
  clampLines?: number;
  /** When set, tapping the text navigates here (e.g. the thread permalink). */
  href?: string;
};

const READ_MORE_COLOR = "#3c839f";

function stripHtml(input: string): string {
  if (!/<[a-z][\s\S]*>/i.test(input)) return input;
  if (typeof window === "undefined") {
    // SSR fallback: turn block/br tags into newlines, drop the rest.
    return input
      .replace(/<\/(p|div|li)>/gi, "\n")
      .replace(/<br\s*\/?>/gi, "\n")
      .replace(/<[^>]+>/g, "")
      .replace(/\n{3,}/g, "\n\n")
      .trim();
  }
  const doc = new DOMParser().parseFromString(input, "text/html");
  doc.querySelectorAll("p, br, div, li").forEach((el) => {
    el.append("\n");
  });
  return (doc.body.textContent ?? "").replace(/\n{3,}/g, "\n\n").trim();
}

export function ExpandableText({ text: raw, fallback, clampLines = 3, href }: Props) {
  const text = useMemo(() => {
    const stripped = stripHtml(raw);
    return stripped || (fallback ?? "").trim();
  }, [raw, fallback]);

  const [expanded, setExpanded] = useState(false);
  const [isOverflowing, setIsOverflowing] = useState(false);
  const ref = useRef<HTMLParagraphElement>(null);

  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;

    // Overflow = the full scroll height exceeds the clamped visible height.
    // Measure with the clamp applied, so this must run while collapsed.
    const measure = () => {
      if (expanded) return;
      setIsOverflowing(el.scrollHeight - el.clientHeight > 1);
    };

    measure();

    const ro = new ResizeObserver(measure);
    ro.observe(el);
    return () => ro.disconnect();
  }, [text, clampLines, expanded]);

  if (!text) return null;

  const clampStyle = expanded
    ? {}
    : {
        display: "-webkit-box",
        WebkitLineClamp: clampLines,
        WebkitBoxOrient: "vertical" as const,
        overflow: "hidden",
      };

  const paragraph = (
    <styled.p
      ref={ref}
      color="fg.default"
      fontWeight="medium"
      style={{
        margin: 0,
        fontSize: "14px",
        lineHeight: "20px",
        whiteSpace: "pre-wrap",
        wordBreak: "break-word",
        cursor: href ? "pointer" : "default",
        ...clampStyle,
      }}
    >
      {text}
    </styled.p>
  );

  return (
    <styled.div w="full">
      {href ? (
        <Link href={href} style={{ textDecoration: "none", color: "inherit" }}>
          {paragraph}
        </Link>
      ) : (
        paragraph
      )}

      {isOverflowing && (
        <styled.button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            setExpanded((v) => !v);
          }}
          style={{
            marginTop: "2px",
            background: "transparent",
            border: "none",
            padding: 0,
            cursor: "pointer",
            color: READ_MORE_COLOR,
            fontSize: "10px",
            fontWeight: 700,
            letterSpacing: "0.48px",
            textTransform: "uppercase",
          }}
        >
          {expanded ? "Read less" : "Read more"}
        </styled.button>
      )}
    </styled.div>
  );
}
