"use client";

import { ReactNode } from "react";

import { styled } from "@/styled-system/jsx";

type Props = {
  icon: ReactNode;
  count?: number;
  active?: boolean;
  ariaLabel: string;
  title?: string;
  onClick?: (e: React.MouseEvent<HTMLElement>) => void;
  /** Render as a non-button element (e.g. when nested inside an anchor). */
  as?: "button" | "span";
};

export function ActionPill({
  icon,
  count,
  active = false,
  ariaLabel,
  title,
  onClick,
  as = "button",
}: Props) {
  const Comp = as === "span" ? styled.span : styled.button;
  return (
    <Comp
      type={as === "button" ? "button" : undefined}
      role={as === "span" ? "button" : undefined}
      onClick={onClick}
      display="flex"
      alignItems="center"
      gap="1"
      aria-label={ariaLabel}
      title={title ?? ariaLabel}
      style={{
        backgroundColor: active ? "#404040" : "transparent",
        border: "none",
        cursor: "pointer",
        padding: "6px 14px",
        borderRadius: "96px",
        color: active ? "#ffffff" : "var(--colors-fg-default)",
      }}
    >
      {icon}
      {count !== undefined && (
        <styled.span
          fontWeight="medium"
          fontVariantNumeric="tabular-nums"
          style={{
            color: active ? "#ffffff" : "#404040",
            fontSize: "12px",
            lineHeight: "16px",
            letterSpacing: "0.24px",
          }}
        >
          {count}
        </styled.span>
      )}
    </Comp>
  );
}
