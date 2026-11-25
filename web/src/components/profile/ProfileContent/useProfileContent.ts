import { useEffect, useState } from "react";

import { Account, PublicProfile, ThreadReference } from "src/api/openapi-schema";

import { useThreadList } from "@/api/openapi-client/threads";

export type Props = {
  session?: Account;
  profile: PublicProfile;
};

export function useProfileContent({ session, profile }: Props) {
  const [page, setPage] = useState(1);
  const [pages, setPages] = useState<Record<number, ThreadReference[]>>({});

  const threads = useThreadList({
    author: profile.handle,
    page: page.toString(),
  });

  // Keep each fetched page so Load More appends rather than replacing the list.
  useEffect(() => {
    if (!threads.data) {
      return;
    }

    const fetched = threads.data.threads;

    setPages((prev) =>
      prev[page] === fetched ? prev : { ...prev, [page]: fetched },
    );
  }, [threads.data, page]);

  if (!threads.data) {
    return { ready: false as const, error: threads.error };
  }

  const isSelf = session?.id === profile.id;
  const { results, total_pages, current_page } = threads.data;

  const merged = Object.keys(pages)
    .map(Number)
    .sort((a, b) => a - b)
    .flatMap((p) => pages[p] ?? []);

  return {
    ready: true as const,
    isSelf,
    data: {
      threads: merged.length > 0 ? merged : threads.data.threads,
      total: results,
      hasMore: current_page < total_pages,
    },
    handlers: {
      handleLoadMore: () => setPage((p) => p + 1),
    },
  };
}
