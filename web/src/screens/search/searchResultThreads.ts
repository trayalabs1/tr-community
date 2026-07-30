import {
  DatagraphItemKind,
  DatagraphItemList,
  Thread,
} from "@/api/openapi-schema";

export function searchResultThreads(
  items: DatagraphItemList | undefined,
): Thread[] {
  if (!items) {
    return [];
  }

  return items.reduce<Thread[]>((acc, item) => {
    if (item.kind === DatagraphItemKind.thread) {
      acc.push(item.ref);
    }

    return acc;
  }, []);
}
