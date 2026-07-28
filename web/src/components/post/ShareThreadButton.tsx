"use client";

import { MouseEvent, useCallback } from "react";
import { toast } from "sonner";

import { useCopyToClipboard } from "@/utils/useCopyToClipboard";

import { Share2 } from "lucide-react";

import { ThreadReference } from "@/api/openapi-schema";
import { getPermalinkForThread } from "@/components/thread/utils";
import { useShare } from "@/utils/client";
import { styled } from "@/styled-system/jsx";

import { ActionPill } from "./ActionPill";

const Share2Icon = styled(Share2);

type Props = {
  thread: ThreadReference;
  channelID?: string;
};

export function ShareThreadButton({ thread }: Props) {
  const isSharingEnabled = useShare();
  const [, copyToClipboard] = useCopyToClipboard();
  const permalink = getPermalinkForThread(thread.slug);

  const handleShare = useCallback(
    async (e: MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();

      if (isSharingEnabled) {
        try {
          await navigator.share({
            title: `A post by ${thread.author.name}`,
            url: permalink,
            text: thread.description,
          });
          return;
        } catch {
          // fall through to copy on cancel/failure
        }
      }

      await copyToClipboard(permalink);
      toast.success("Link copied to clipboard");
    },
    [isSharingEnabled, copyToClipboard, permalink, thread.author.name, thread.description],
  );

  return (
    <ActionPill
      ariaLabel="Share post"
      onClick={handleShare}
      icon={<Share2Icon width="4" height="4" />}
    />
  );
}
