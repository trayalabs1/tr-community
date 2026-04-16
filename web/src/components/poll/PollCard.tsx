"use client";

import { css } from "@/styled-system/css";
import { VStack, HStack } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { usePollCard } from "./usePollCard";

type Props = {
  threadMark: string;
  optionDefs: Array<{ id: string; text: string }>;
};

export function PollCard({ threadMark, optionDefs }: Props) {
  const { status, vote } = usePollCard(threadMark);

  const hasVoted = status != null && status.user_vote != null;
  const totalVotes = status?.total_votes ?? 0;

  if (!hasVoted) {
    return (
      <VStack gap="2" w="full" alignItems="start">
        {optionDefs.map((opt) => (
          <button
            key={opt.id}
            type="button"
            onClick={() => vote(opt.id)}
            className={css({
              w: "full",
              p: "3",
              borderRadius: "md",
              borderWidth: "thin",
              borderStyle: "solid",
              borderColor: "border.default",
              textAlign: "left",
              cursor: "pointer",
              fontSize: "sm",
              color: "fg.default",
              bg: "bg.default",
              _hover: { borderColor: "fg.muted" },
            })}
          >
            {opt.text}
          </button>
        ))}
      </VStack>
    );
  }

  const options = status!.options;

  return (
    <VStack gap="2" w="full" alignItems="start">
      {options.map((opt) => {
        const pct = totalVotes > 0 ? (opt.votes / totalVotes) * 100 : 0;
        const isSelected = status!.user_vote === opt.id;

        return (
          <div
            key={opt.id}
            className={css({
              w: "full",
              borderRadius: "md",
              borderWidth: "thin",
              borderStyle: "solid",
              borderColor: "border.default",
              overflow: "hidden",
              position: "relative",
            })}
            style={
              isSelected
                ? { borderColor: TRAYA_COLORS.primary }
                : undefined
            }
          >
            <div
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                height: "100%",
                width: `${pct}%`,
                background: isSelected
                  ? TRAYA_COLORS.tertiary
                  : "var(--colors-bg-muted)",
                transition: "width 0.3s ease",
              }}
            />
            <HStack
              justify="space-between"
              px="3"
              py="2"
              position="relative"
              style={{ zIndex: 1 }}
            >
              <span
                className={css({ fontSize: "sm", color: "fg.default" })}
                style={isSelected ? { fontWeight: 600 } : undefined}
              >
                {opt.text}
              </span>
              <span
                className={css({ fontSize: "sm", color: "fg.muted" })}
              >
                {opt.votes}
              </span>
            </HStack>
          </div>
        );
      })}

      <p
        className={css({ fontSize: "xs", color: "fg.muted", mt: "1" })}
      >
        {totalVotes} {totalVotes === 1 ? "vote" : "votes"}
      </p>
    </VStack>
  );
}
