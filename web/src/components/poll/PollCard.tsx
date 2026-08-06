"use client";

import { Check } from "lucide-react";

import { VStack, HStack, styled } from "@/styled-system/jsx";
import { usePollCard } from "./usePollCard";

type Props = {
  threadMark: string;
  optionDefs: Array<{ id: string; text: string }>;
};

const BORDER = "#dedede";
const TEXT = "#404040";
const MUTED = "#999999";

function OptionRow({
  label,
  count,
  checked,
  onClick,
}: {
  label: string;
  count?: number;
  checked: boolean;
  onClick?: () => void;
}) {
  return (
    <styled.button
      type="button"
      onClick={onClick}
      display="flex"
      alignItems="center"
      gap="2"
      w="full"
      bg="white"
      cursor={onClick ? "pointer" : "default"}
      style={{
        border: `1px solid ${BORDER}`,
        borderRadius: "12px",
        padding: "12px 16px",
      }}
    >
      <styled.span
        display="flex"
        alignItems="center"
        justifyContent="center"
        flexShrink="0"
        style={{
          width: "16px",
          height: "16px",
          borderRadius: "4px",
          border: `1.5px solid ${checked ? TEXT : BORDER}`,
          backgroundColor: checked ? TEXT : "transparent",
        }}
      >
        {checked && <Check size={11} color="white" strokeWidth={3} />}
      </styled.span>
      <styled.span
        flex="1"
        textAlign="left"
        style={{
          color: TEXT,
          fontSize: "14px",
          fontWeight: 500,
          lineHeight: "20px",
        }}
      >
        {label}
      </styled.span>
      {count !== undefined && (
        <styled.span
          style={{
            color: TEXT,
            fontSize: "14px",
            fontWeight: 500,
            lineHeight: "20px",
          }}
        >
          {count}
        </styled.span>
      )}
    </styled.button>
  );
}

export function PollCard({ threadMark, optionDefs }: Props) {
  const { status, vote } = usePollCard(threadMark);

  const hasVoted = status != null && status.user_vote != null;

  return (
    <VStack gap="6" w="full" alignItems="start">
      <HStack gap="1.5" alignItems="center">
        <styled.span
          display="flex"
          alignItems="center"
          justifyContent="center"
          flexShrink="0"
          style={{
            width: "18px",
            height: "18px",
            borderRadius: "1000px",
            backgroundColor: "#b5b5b5",
          }}
        >
          <Check size={11} color="white" strokeWidth={3} />
        </styled.span>
        <styled.span
          style={{
            color: MUTED,
            fontSize: "12px",
            fontWeight: 600,
            lineHeight: "16px",
            letterSpacing: "0.24px",
          }}
        >
          Select an option
        </styled.span>
      </HStack>

      <VStack gap="3" w="full" alignItems="start">
        {hasVoted
          ? status!.options.map((opt) => (
              <OptionRow
                key={opt.id}
                label={opt.text}
                count={opt.votes}
                checked={status!.user_vote === opt.id}
              />
            ))
          : optionDefs.map((opt) => (
              <OptionRow
                key={opt.id}
                label={opt.text}
                checked={false}
                onClick={() => vote(opt.id)}
              />
            ))}
      </VStack>
    </VStack>
  );
}
