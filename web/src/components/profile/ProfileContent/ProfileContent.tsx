import { Unready } from "src/components/site/Unready";

import { ThreadReferenceList } from "@/components/post/ThreadReferenceList";
import * as Tabs from "@/components/ui/tabs";
import { VStack } from "@/styled-system/jsx";

import { Props, useProfileContent } from "./useProfileContent";

export function ProfileContent(props: Props) {
  const { ready, error, data } = useProfileContent(props);

  if (!ready) {
    return <Unready error={error} />;
  }

  const { threads } = data;

  return (
    <VStack alignItems="start" w="full" gap="0">
      <Tabs.Root width="full" variant="line" defaultValue="posts">
        <Tabs.List>
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

        <Tabs.Content value="posts">
          <ThreadReferenceList threads={threads} />
        </Tabs.Content>
      </Tabs.Root>
    </VStack>
  );
}
