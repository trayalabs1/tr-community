"use client";

import { DatagraphSearchResults } from "src/components/search/DatagraphSearchResults";
import { UnreadyBanner } from "src/components/site/Unready";

import { PaginationControls } from "@/components/site/PaginationControls/PaginationControls";
import { Button } from "@/components/ui/button";
import { CancelIcon } from "@/components/ui/icons/Cancel";
import { SearchIcon } from "@/components/ui/icons/Search";
import { Input } from "@/components/ui/input";
import { styled } from "@/styled-system/jsx";

import { Props, useSearchScreen } from "./useSearch";

export function SearchScreen(props: Props) {
  const { form, error, isLoading, data, handlers } = useSearchScreen(props);

  const { query, page, results } = data;

  return (
    <styled.div
      display="flex"
      flexDirection="column"
      h="full"
      bg="bg.site"
      pb="20"
    >
      <styled.div
        position="sticky"
        top="0"
        bg="bg.site"
        backdropFilter="auto"
        px="4"
        py="2"
        css={{
          zIndex: 10,
          borderBottom: "1px solid var(--colors-border-default)",
          "@media (max-width: 640px)": {
            marginLeft: "calc(-100vw + 100%)",
            marginRight: "calc(-100vw + 100%)",
            paddingLeft: "calc(100vw - 100%)",
            paddingRight: "calc(100vw - 100%)",
          },
        } as any}
      >
        <styled.h1
          fontWeight="bold"
          fontSize="xl"
          mb="2"
          px="2"
        >
          Search
        </styled.h1>

        <styled.form
          display="flex"
          gap="2"
          onSubmit={handlers.handleSearch}
          action="/search"
        >
          <styled.div
            position="relative"
            flex="1"
            display="flex"
            alignItems="center"
            gap="0"
            px="2"
            py="1"
            borderRadius="full"
            bg="bg.searchInput"
            css={{
              border: "1.5px solid transparent",
              transition: "border-color 0.2s ease",
              "&:has(input:focus)": {
                borderColor: "#2D7A4A",
              },
            } as any}
          >
            <styled.div
              position="absolute"
              left="3"
              display="flex"
              alignItems="center"
              pointerEvents="none"
              css={{
                color: "#a2c3ac",
              } as any}
            >
              <SearchIcon />
            </styled.div>

            <Input
              flex="1"
              pl="11"
              pr={query ? "10" : "3"}
              type="search"
              defaultValue={props.initialQuery}
              placeholder="Search all posts..."
              _focus={{
                boxShadow: "none" as any,
                outline: "none",
              }}
              css={{
                width: "100%",
                overflow: "hidden",
                background: "transparent",
                border: "none",
                borderRadius: "0",
              } as any}
              {...form.register("q")}
            />

            {query && (
              <Button
                position="absolute"
                right="3"
                size="sm"
                variant="ghost"
                type="reset"
                onClick={handlers.handleReset}
                p="1"
              >
                <CancelIcon />
              </Button>
            )}
          </styled.div>
        </styled.form>
      </styled.div>

      {/* Filter tags - commented out as per design */}
      {/* <styled.div
        w="full"
        p="4"
        css={{
          display: "flex",
          borderBottom: "1px solid var(--colors-border-default)",
        } as any}
      >
        <DatagraphKindFilterField
          control={form.control}
          name="kind"
          items={[
            {
              label: "Threads",
              description: "Include discussion threads in the search.",
              icon: <DiscussionIcon />,
              value: DatagraphItemKind.thread,
            },
            {
              label: "Replies",
              description:
                "Include replies to discussion threads in the search.",
              icon: <ReplyIcon />,
              value: DatagraphItemKind.reply,
            },
            {
              label: "Library",
              description: "Include library pages in the search.",
              icon: <LibraryIcon />,
              value: DatagraphItemKind.node,
            },
          ]}
        />
      </styled.div> */}

      <styled.div
        flex="1"
        overflowY="auto"
        p="4"
        display="flex"
        flexDirection="column"
        gap="4"
      >
        {!query ? (
          <styled.div
            display="flex"
            flexDirection="column"
            alignItems="center"
            justifyContent="center"
            py="16"
            textAlign="center"
          >
            <styled.div
              w="16"
              h="16"
              borderRadius="full"
              display="flex"
              alignItems="center"
              justifyContent="center"
              mb="4"
              css={{
                backgroundColor: "#f0f5f1",
                color: "#a2c3ac",
              } as any}
            >
              <SearchIcon />
            </styled.div>
            <styled.p
              css={{
                color: "var(--colors-text-secondary)",
              }}
            >
              Search for posts across all channels
            </styled.p>
          </styled.div>
        ) : results ? (
          <>
            {results?.items.length > 0 ? (
              <>
                <styled.p
                  fontSize="sm"
                  mb="2"
                  css={{
                    color: "var(--colors-text-secondary)",
                  }}
                >
                  {results.items.length} result{results.items.length !== 1 ? "s" : ""} found
                </styled.p>

                <PaginationControls
                  path="/search"
                  params={{ q: query }}
                  currentPage={page}
                  totalPages={results.total_pages}
                  pageSize={results.page_size}
                />

                <DatagraphSearchResults result={results} />
              </>
            ) : (
              <styled.div textAlign="center" py="16">
                <styled.p
                  css={{
                    color: "var(--colors-text-secondary)",
                  }}
                >
                  No posts found for "{query}"
                </styled.p>
              </styled.div>
            )}
          </>
        ) : (
          isLoading && <UnreadyBanner error={error} />
        )}
      </styled.div>
    </styled.div>
  );
}
