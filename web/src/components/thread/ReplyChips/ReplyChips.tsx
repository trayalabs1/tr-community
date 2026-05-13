"use client";

import { useMemo } from "react";

import { Button } from "@/components/ui/button";
import { HStack, styled } from "@/styled-system/jsx";

type Props = {
  candidates: string[];
  onPick: (text: string) => void;
  isPosting?: boolean;
  posted?: boolean;
};

const MAX_CHIPS = 2;

export function ReplyChips({ candidates, onPick, isPosting, posted }: Props) {
  const shown = useMemo(() => pickRandom(candidates, MAX_CHIPS), [candidates]);

  if (shown.length === 0) {
    return null;
  }

  if (posted) {
    return (
      <styled.div width="full" py="1">
        <styled.span fontSize="xs" color="fg.muted">
          Reply posted
        </styled.span>
      </styled.div>
    );
  }

  return (
    <styled.div
      width="full"
      overflowX="auto"
      style={{ scrollbarWidth: "none" }}
      css={{ "&::-webkit-scrollbar": { display: "none" } }}
    >
      <HStack gap="2" whiteSpace="nowrap" py="1">
        {shown.map((text) => (
          <Button
            key={text}
            type="button"
            variant="outline"
            size="xs"
            disabled={isPosting}
            onClick={() => onPick(text)}
          >
            {text}
          </Button>
        ))}
      </HStack>
    </styled.div>
  );
}

function pickRandom<T>(items: T[], count: number): T[] {
  if (items.length <= count) {
    return items.slice();
  }
  const pool = items.slice();
  const out: T[] = [];
  for (let i = 0; i < count; i++) {
    const idx = Math.floor(Math.random() * pool.length);
    out.push(pool[idx]!);
    pool.splice(idx, 1);
  }
  return out;
}
