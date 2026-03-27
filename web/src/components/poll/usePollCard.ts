"use client";

import { useState, useEffect, useCallback } from "react";
import { threadGetPoll, threadVotePoll } from "@/api/openapi-client/threads";
import type { PollStatus } from "@/api/openapi-schema";

export function usePollCard(threadMark: string) {
  const [status, setStatus] = useState<PollStatus | null>(null);

  useEffect(() => {
    threadGetPoll(threadMark).then(setStatus).catch(() => null);
  }, [threadMark]);

  const vote = useCallback(
    async (optionId: string) => {
      const updated = await threadVotePoll(threadMark, {
        option_id: optionId,
      });
      setStatus(updated);
    },
    [threadMark],
  );

  return { status, vote };
}
