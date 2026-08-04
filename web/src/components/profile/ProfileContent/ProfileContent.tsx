import { Unready } from "src/components/site/Unready";

import { ThreadReferenceList } from "@/components/post/ThreadReferenceList";
import * as Tabs from "@/components/ui/tabs";
import { VStack, styled } from "@/styled-system/jsx";

import { Props, useProfileContent } from "./useProfileContent";

export function ProfileContent(props: Props) {
  const result = useProfileContent(props);

  if (!result.ready) {
    return <Unready error={result.error} />;
  }

  const { isSelf, data, handlers } = result;
  const { threads, total, hasMore } = data;

  const list = (
    <>
      <ThreadReferenceList threads={threads} />

      {hasMore && (
        <styled.div w="full" px="4" py="4" bg="bg.savedPage">
          <styled.button
            type="button"
            onClick={handlers.handleLoadMore}
            display="flex"
            alignItems="center"
            justifyContent="center"
            w="full"
            h="12"
            borderRadius="[12px]"
            fontSize="sm"
            fontWeight="semibold"
            color="fg.default"
            bg="bg.surfaceWhite"
            borderWidth="thin"
            borderStyle="solid"
            borderColor="border.default"
            cursor="pointer"
          >
            Load More
          </styled.button>
        </styled.div>
      )}
    </>
  );

  if (!isSelf) {
    return (
      <VStack alignItems="start" w="full" gap="0">
        <styled.div w="full" px="4" py="3" bg="bg.savedPage">
          <styled.h2
            fontSize="md"
            lineHeight="[24px]"
            fontWeight="semibold"
            color="fg.default"
          >
            Threads ({total})
          </styled.h2>
        </styled.div>

        {list}
      </VStack>
    );
  }

  return (
    <VStack alignItems="start" w="full" gap="0">
      <Tabs.Root width="full" variant="line" defaultValue="posts">
        <Tabs.List justifyContent="center">
          <Tabs.Trigger
            value="posts"
            textTransform="uppercase"
            letterSpacing="[0.4px]"
            fontWeight="medium"
          >
            Posts
          </Tabs.Trigger>
          <Tabs.Indicator />
        </Tabs.List>

        <Tabs.Content value="posts">{list}</Tabs.Content>
      </Tabs.Root>
    </VStack>
  );
}
