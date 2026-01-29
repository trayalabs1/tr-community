import { useRouter } from "next/navigation";

import { useCollectionList } from "src/api/openapi-client/collections";
import { Account, Collection, PostReference } from "src/api/openapi-schema";

import { handle } from "@/api/client";
import { useCollectionItemMutations } from "@/lib/collection/mutation";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

export type Props = {
  account: Account;
  thread: PostReference;
};

export function useCollectionMenu({ account, thread }: Props) {
  const router = useRouter();
  const { data, error } = useCollectionList({
    account_handle: account.handle,
    has_item: thread.id,
  });

  const { addPost, removePost, revalidate } =
    useCollectionItemMutations(account, router);
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
