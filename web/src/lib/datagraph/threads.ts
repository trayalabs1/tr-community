import {
  DatagraphItemKind,
  DatagraphItemList,
  Thread,
} from "@/api/openapi-schema";

export function datagraphItemThreads(
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

// Post lacks ThreadReferenceProps (reply_status, tags, pinned, channel_id).
// ThreadReferenceCard reads those defensively, so a Post renders without a
// reply count or category badge rather than throwing.
export function datagraphItemPostsAndThreads(
  items: DatagraphItemList | undefined,
): Thread[] {
  if (!items) {
    return [];
  }

  return items.reduce<Thread[]>((acc, item) => {
    if (
      item.kind === DatagraphItemKind.thread ||
      item.kind === DatagraphItemKind.post
    ) {
      acc.push(item.ref as Thread);
    }

    return acc;
  }, []);
}
