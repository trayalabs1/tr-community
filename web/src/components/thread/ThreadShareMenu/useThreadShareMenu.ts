import { useState } from "react";

import { handle } from "@/api/client";
import { channelThreadCreate } from "@/api/openapi-client/channels";
import { ThreadReference } from "@/api/openapi-schema";
import { useFeedMutations } from "@/lib/feed/mutation";

type Props = {
  thread: ThreadReference;
};

export function useThreadShareMenu({ thread }: Props) {
  const { revalidate } = useFeedMutations();

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [subtitle, setSubtitle] = useState("");

  function handleToggleChannel(channelID: string) {
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(channelID)) {
        next.delete(channelID);
      } else {
        next.add(channelID);
      }
      return next;
    });
  }

  function handleSubtitleChange(event: React.ChangeEvent<HTMLInputElement>) {
    setSubtitle(event.target.value);
  }

  async function handleConfirmShare() {
    const destinations = Array.from(selected);

    if (destinations.length === 0) return;

    const title = thread.title || "Untitled post";

    await handle(
      async () => {
        await Promise.all(
          destinations.map((channelID) =>
            channelThreadCreate(channelID, {
              title,
              body: subtitle.trim(),
              reference_post_id: thread.id,
              visibility: "published",
            }),
          ),
        );

        setSelected(new Set());
        setSubtitle("");
      },
      {
        promiseToast: {
          loading: "Sharing thread...",
          success: `Shared to ${destinations.length} channel${destinations.length === 1 ? "" : "s"}!`,
        },
        cleanup: async () => {
          await revalidate();
        },
      },
    );
  }

  return {
    selected,
    subtitle,
    handlers: {
      handleToggleChannel,
      handleSubtitleChange,
      handleConfirmShare,
    },
  };
}
