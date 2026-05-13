"use client";

import { useState } from "react";

import { handle } from "@/api/client";
import { ThreadReference } from "@/api/openapi-schema";
import { useFeedMutations } from "@/lib/feed/mutation";
import { useThreadMutations } from "@/lib/thread/mutation";

export function usePostReplyChip(thread: ThreadReference) {
  const { createReply } = useThreadMutations(thread);
  const { revalidate: revalidateFeed } = useFeedMutations();
  const [isPosting, setPosting] = useState(false);
  const [posted, setPosted] = useState(false);

  async function postChip(text: string) {
    if (isPosting || posted) return;
    setPosting(true);
    try {
      await handle(async () => {
        await createReply({ body: `<p>${escapeHtml(text)}</p>` });
        await revalidateFeed();
        setPosted(true);
      });
    } finally {
      setPosting(false);
    }
  }

  return { postChip, isPosting, posted };
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}
