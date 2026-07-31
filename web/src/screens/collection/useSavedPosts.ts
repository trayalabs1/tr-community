"use client";

import {
  useCollectionGet,
  useCollectionList,
} from "@/api/openapi-client/collections";
import { CollectionListOKResponse } from "@/api/openapi-schema";
import { datagraphItemPostsAndThreads } from "@/lib/datagraph/threads";

export type Props = {
  initialCollections?: CollectionListOKResponse;
};

export function useSavedPosts(props: Props) {
  const {
    data: list,
    error: listError,
    isLoading: isListLoading,
  } = useCollectionList(undefined, {
    swr: { fallbackData: props.initialCollections },
  });

  const defaultCollection =
    list?.collections.find((c) => c.is_default) ?? list?.collections[0];

  const {
    data: collection,
    error: collectionError,
    isLoading: isCollectionLoading,
  } = useCollectionGet(defaultCollection?.slug ?? "", {
    swr: { enabled: Boolean(defaultCollection?.slug) },
  });

  const threads = datagraphItemPostsAndThreads(
    collection?.items.map((i) => i.item),
  );

  return {
    isLoading: isListLoading || isCollectionLoading,
    error: listError ?? collectionError,
    hasCollection: Boolean(defaultCollection),
    threads,
  };
}
