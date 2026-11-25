"use client";

import { CollectionListOKResponse } from "@/api/openapi-schema";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { CenteredBackHeader } from "@/components/site/Header";
import { UnreadyBanner } from "@/components/site/Unready";
import { styled } from "@/styled-system/jsx";

import { useSavedPosts } from "./useSavedPosts";

export type Props = {
  initialCollections?: CollectionListOKResponse;
};

export function SavedPostsScreen(props: Props) {
  const { isLoading, error, threads } = useSavedPosts(props);

  return (
    <styled.div display="flex" flexDirection="column" w="full" h="full">
      <CenteredBackHeader title="Saved Posts" />

      <styled.div
        flex="1"
        display="flex"
        flexDirection="column"
        w="full"
        maxW="2xl"
        mx="auto"
        py="4"
        bg="bg.savedPage"
      >
        {threads.length > 0 ? (
          <styled.ol
            display="flex"
            flexDirection="column"
            gap="2"
            w="full"
            p="0"
            m="0"
          >
            {threads.map((thread) => (
              <styled.li key={thread.id} listStyleType="none" w="full">
                <ThreadReferenceCard thread={thread} />
              </styled.li>
            ))}
          </styled.ol>
        ) : isLoading ? (
          <UnreadyBanner error={error} />
        ) : (
          <styled.div
            display="flex"
            alignItems="center"
            justifyContent="center"
            w="full"
            py="16"
            px="4"
          >
            <styled.p fontSize="sm" color="fg.muted" textAlign="center">
              No saved posts yet.
            </styled.p>
          </styled.div>
        )}
      </styled.div>
    </styled.div>
  );
}
