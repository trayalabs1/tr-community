"use client";

import { useState } from "react";

import { handle } from "@/api/client";
import { Permission, ThreadReference } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { useFeedMutations } from "@/lib/feed/mutation";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { useThreadMutations } from "@/lib/thread/mutation";
import { hasPermission } from "@/utils/permissions";

export function usePostReplyChip(thread: ThreadReference) {
  const session = useSession();
  const { createReply } = useThreadMutations(thread);
  const { revalidate: revalidateFeed } = useFeedMutations();
  const { trackCardReply, trackAdminReplied } = useEventTracking();
  const [isPosting, setPosting] = useState(false);
  const [posted, setPosted] = useState(false);

  async function postChip(text: string) {
    if (isPosting || posted) return;
    setPosting(true);

    const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);

    trackCardReply(thread.id, text.length, undefined, "quick_chips");

    if (isAdmin) {
      trackAdminReplied(thread.id, text.length, thread.author.id, undefined, "quick_chips");
    }

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
