import { DatagraphSearchOKResponse } from "@/api/openapi-schema";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { PaginationControls } from "@/components/site/PaginationControls/PaginationControls";
import { styled } from "@/styled-system/jsx";

import { searchResultThreads } from "./searchResultThreads";

type Props = {
  results: DatagraphSearchOKResponse;
  query: string;
  page: number;
};

export function SearchResults({ results, query, page }: Props) {
  const threads = searchResultThreads(results.items);

  if (threads.length === 0) {
    return (
      <styled.div display="flex" justifyContent="center" py="8" px="4">
        <styled.p
          fontSize="sm"
          lineHeight="[20px]"
          color="fg.muted"
          textAlign="center"
        >
          No posts found for &ldquo;{query}&rdquo;
        </styled.p>
      </styled.div>
    );
  }

  return (
    <styled.div display="flex" flexDirection="column" gap="3">
      <styled.p fontSize="sm" lineHeight="[20px]" color="fg.muted">
        {results.results} result{results.results === 1 ? "" : "s"} found
      </styled.p>

      <styled.ol display="flex" flexDirection="column" gap="3" w="full">
        {threads.map((thread) => (
          <ThreadReferenceCard key={thread.id} thread={thread} />
        ))}
      </styled.ol>

      <PaginationControls
        path="/search"
        params={{ q: query }}
        currentPage={page}
        totalPages={results.total_pages}
        pageSize={results.page_size}
      />
    </styled.div>
  );
}
