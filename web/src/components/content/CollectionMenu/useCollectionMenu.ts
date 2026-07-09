import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { collectionList } from "src/api/openapi-client/collections";
import { Account, PostReference } from "src/api/openapi-schema";

import { handle } from "@/api/client";
import { useCollectionItemMutations } from "@/lib/collection/mutation";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

export type Props = {
  account: Account;
  thread: PostReference;
};

export function useToggleSave({ account, thread }: Props) {
  const router = useRouter();
  const [isSaved, setIsSaved] = useState(thread.collections.has_collected);

  useEffect(() => {
    setIsSaved(thread.collections.has_collected);
  }, [thread.collections.has_collected]);

  const { addPostToDefault, removePost, revalidate } =
    useCollectionItemMutations(account, router);
  const { trackCardSave } = useEventTracking();

  const onToggle = async () => {
    const next = !isSaved;
    setIsSaved(next);
    trackCardSave(thread.id, next ? "save" : "unsave", undefined);

    await handle(
      async () => {
        if (next) {
          await addPostToDefault(thread.id);
        } else {
          const { collections } = await collectionList({
            account_handle: account.handle,
            has_item: thread.id,
          });

          await Promise.all(
            collections
              .filter((c) => c.has_queried_item)
              .map((c) => removePost(c.id, thread.id)),
          );
        }
      },
      {
        onError: async () => setIsSaved(!next),
        cleanup: async () => await revalidate(),
      },
    );
  };

  return {
    isSaved,
    onToggle,
  };
}
