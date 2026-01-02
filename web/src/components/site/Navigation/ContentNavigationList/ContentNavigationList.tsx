"use client";

import {
  CategoryListOKResponse,
  ChannelListOKResponse,
  NodeListResult,
} from "@/api/openapi-schema";
import { useChannelCategoryList } from "@/api/openapi-client/channels";
import { CategoryList } from "@/components/category/CategoryList/CategoryList";
import { ChannelList } from "@/components/channel/ChannelList/ChannelList";
import { LStack, styled } from "@/styled-system/jsx";

import { CollectionsAnchor } from "../Anchors/Collections";
import { LinksAnchor } from "../Anchors/Link";
import { MembersAnchor } from "../Anchors/Members";
import { LibraryNavigationTree } from "../LibraryNavigationTree/LibraryNavigationTree";
import { useNavigation } from "../useNavigation";

type Props = {
  initialNodeList?: NodeListResult;
  initialCategoryList?: CategoryListOKResponse;
  initialChannelList?: ChannelListOKResponse;
};

export function ContentNavigationList(props: Props) {
  const { nodeSlug, isChannelPage, channelID } = useNavigation();

  // Fetch channel-specific categories if on a channel page
  const { data: channelCategories } = useChannelCategoryList(channelID ?? "", {
    swr: {
      enabled: isChannelPage && !!channelID,
    },
  });

  return (
    <styled.nav
      display="flex"
      flexDir="column"
      gap="4"
      height="full"
      width="full"
      minH="0"
      alignItems="start"
      justifyContent="space-between"
    >
      <LStack
        gap="1"
        overflowY="scroll"
        style={{
          scrollbarWidth: "none",
        }}
      >
        <ChannelList
          initialChannelList={props.initialChannelList}
          currentChannelID={channelID}
          channelCategories={channelCategories}
          channelNodes={props.initialNodeList}
        />
        {!isChannelPage && (
          <CategoryList
            initialCategoryList={props.initialCategoryList}
            channelID={undefined}
          />
        )}
        <LibraryNavigationTree
          initialNodeList={props.initialNodeList}
          currentNode={nodeSlug}
          visibility={["draft", "review", "unlisted", "published"]}
        />
      </LStack>

      <LStack gap="1">
        <CollectionsAnchor />
        <LinksAnchor />
        <MembersAnchor />
      </LStack>
    </styled.nav>
  );
}
