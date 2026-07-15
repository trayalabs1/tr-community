import { useEffect, useRef, useState } from "react";

import { handle } from "@/api/client";
import {
  threadShareCreate,
  threadShareDelete,
  useThreadShareList,
} from "@/api/openapi-client/threads";
import { ThreadReference } from "@/api/openapi-schema";
import { useFeedMutations } from "@/lib/feed/mutation";

type Props = {
  thread: ThreadReference;
  // Only fetch the current share state once the submenu is opened, to avoid a
  // request per thread in the feed.
  enabled: boolean;
};

export function useThreadShareMenu({ thread, enabled }: Props) {
  const { revalidate } = useFeedMutations();

  const { data, mutate: mutateShares } = useThreadShareList(thread.slug, {
    swr: { enabled },
  });

  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [subtitle, setSubtitle] = useState("");

  // Seed the selection with the already-shared channels once they load, so the
  // menu opens reflecting the thread's current share state.
  const seededRef = useRef(false);
  useEffect(() => {
    if (seededRef.current || !data) return;
    seededRef.current = true;
    setSelected(new Set(data.shares.map((s) => s.channel.id)));
  }, [data]);

  const initialShared = new Set((data?.shares ?? []).map((s) => s.channel.id));

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
    const toShare = Array.from(selected).filter((id) => !initialShared.has(id));
    const toUnshare = Array.from(initialShared).filter(
      (id) => !selected.has(id),
    );

    if (toShare.length === 0 && toUnshare.length === 0) return;

    await handle(
      async () => {
        if (toShare.length > 0) {
          await threadShareCreate(thread.slug, {
            channels: toShare,
            subtitle: subtitle.trim() || undefined,
          });
        }

        await Promise.all(
          toUnshare.map((id) => threadShareDelete(thread.slug, id)),
        );

        setSubtitle("");
      },
      {
        promiseToast: {
          loading: "Updating shares...",
          success: "Shares updated!",
        },
        cleanup: async () => {
          await mutateShares();
          await revalidate();
        },
      },
    );
  }

  return {
    selected,
    subtitle,
    initialShared,
    handlers: {
      handleToggleChannel,
      handleSubtitleChange,
      handleConfirmShare,
    },
  };
}
