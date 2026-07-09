import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import {
  collectionList,
  useCollectionList,
} from "src/api/openapi-client/collections";
import { Account, Collection, PostReference } from "src/api/openapi-schema";

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

export function useQuickSave({ account, thread }: Props) {
  const router = useRouter();
  const { addPostToDefault, revalidate } = useCollectionItemMutations(
    account,
    router,
  );
  const { trackCardSave } = useEventTracking();

  return async () => {
    if (thread.collections.has_collected) {
      return;
    }

    trackCardSave(thread.id, "save", undefined);

    await handle(
      async () => {
        await addPostToDefault(thread.id);
      },
      { cleanup: async () => await revalidate() },
    );
  };
}

export function useCollectionMenu({ account, thread }: Props) {
  const router = useRouter();
  const { data, error } = useCollectionList({
    account_handle: account.handle,
    has_item: thread.id,
  });

  const { addPost, removePost, revalidate } = useCollectionItemMutations(
    account,
    router,
  );
  const { trackCardSave } = useEventTracking();

  if (!data) {
    return {
      ready: false as const,
      error,
    };
  }

  const { collections } = data;

  const handleSelect = (collection: Collection) => async () => {
    const isAlreadySavedIn = collection?.has_queried_item;
    trackCardSave(thread.id, isAlreadySavedIn ? "unsave" : "save", undefined);

    await handle(
      async () => {
        if (isAlreadySavedIn) {
          await removePost(collection.id, thread.id);
        } else {
          await addPost(collection, thread.id);
        }
      },
      { cleanup: async () => await revalidate() },
    );
  };

  return {
    ready: true as const,
    collections,
    handleSelect,
  };
}
