"use client";

import { useRouter } from "next/navigation";

import { UnreadyBanner } from "@/components/site/Unready";
import { styled } from "@/styled-system/jsx";

import { SearchEmptyState } from "./SearchEmptyState";
import { SearchHeader } from "./SearchHeader";
import { SearchResults } from "./SearchResults";
import { Props, useSearchScreen } from "./useSearch";

export function SearchScreen(props: Props) {
  const { form, error, isLoading, data, handlers } = useSearchScreen(props);
  const router = useRouter();

  const { query, page, results } = data;

  return (
    <styled.div display="flex" flexDirection="column" h="full" bg="bg.site">
      <SearchHeader
        register={form.register("q")}
        hasQuery={Boolean(query)}
        onClear={handlers.handleReset}
        onSubmit={handlers.handleSearch}
        onClose={() => router.back()}
      />

      <styled.div flex="1" w="full" maxW="2xl" mx="auto" px="4" pt="4" pb="20">
        {!query ? (
          <SearchEmptyState />
        ) : results ? (
          <SearchResults results={results} query={query} page={page} />
        ) : isLoading ? (
          <UnreadyBanner error={error} />
        ) : null}
      </styled.div>
    </styled.div>
  );
}
