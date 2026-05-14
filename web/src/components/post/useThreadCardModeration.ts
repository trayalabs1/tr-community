import { useRouter } from "next/navigation";
import { useState } from "react";

import { handle } from "@/api/client";
import { ThreadReference } from "@/api/openapi-schema";
import { useConfirmation } from "@/components/site/useConfirmation";
import { useFeedMutations } from "@/lib/feed/mutation";
import { withUndo } from "@/lib/thread/undo";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

export function useThreadCardModeration(thread: ThreadReference) {
  const router = useRouter();
  const { updateThread, deleteThread, revalidate } = useFeedMutations(undefined, undefined, undefined, router);
  const { trackAdminApproved } = useEventTracking();
  const [isDismissing, setIsDismissing] = useState(false);

  const {
    isConfirming: isConfirmingDelete,
    handleConfirmAction: handleConfirmDelete,
    handleCancelAction: handleCancelDelete,
  } = useConfirmation(handleDelete);

  async function handleAcceptThread() {
    trackAdminApproved(thread.id, thread.author.id, thread.channel_id);

    await handle(
      async () => {
        await updateThread(thread.id, { visibility: "published" });
      },
      {
        promiseToast: {
          loading: "Accepting...",
          success: "Thread accepted!",
        },
        cleanup: async () => {
          await revalidate();
        },
      },
    );
  }

  function handleEditAndAccept() {
    router.push(`/t/${thread.slug}?edit=true`);
  }

  async function handleDelete() {
    await handle(async () => {
      await withUndo({
        message: "Thread deleted",
        duration: 5000,
        toastId: `thread-${thread.id}`,
        action: async () => {
          await deleteThread(thread.id);
        },
        onUndo: () => {},
      });
    });
  }

  async function handleDismissThread() {
    if (isDismissing) return;
    setIsDismissing(true);
    try {
      await handle(async () => {
        await updateThread(thread.id, { visibility: "archived" });
        await revalidate();

        await withUndo({
          message: "Thread dismissed",
          duration: 5000,
          toastId: `thread-${thread.id}`,
          action: async () => {},
          onUndo: async () => {
            await updateThread(thread.id, { visibility: "review" });
            await revalidate();
          },
        });
      });
    } finally {
      setIsDismissing(false);
    }
  }

  return {
    isConfirmingDelete,
    isDismissing,
    handlers: {
      handleAcceptThread,
      handleEditAndAccept,
      handleConfirmDelete,
      handleCancelDelete,
      handleDismissThread,
    },
  };
}
