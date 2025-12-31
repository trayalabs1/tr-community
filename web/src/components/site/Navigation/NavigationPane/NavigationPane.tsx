import {
  CategoryListOKResponse,
  ChannelListOKResponse,
  NodeListResult,
} from "@/api/openapi-schema";
import { categoryList } from "@/api/openapi-server/categories";
import { channelList } from "@/api/openapi-server/channels";
import { nodeList } from "@/api/openapi-server/nodes";
import { Box, styled } from "@/styled-system/jsx";
import { Floating } from "@/styled-system/patterns";

import { ContentNavigationList } from "../ContentNavigationList/ContentNavigationList";

import { AdminZone } from "./AdminZone/AdminZone";

export async function NavigationPane() {
  try {
    const { data: initialNodeList } = await nodeList({
      // NOTE: This doesn't work due to a bug in Orval.
      // visibility: ["draft", "review", "unlisted", "published"],
    });
    const { data: initialCategoryList } = await categoryList();
    const { data: initialChannelList } = await channelList({});

    return (
      <NavigationPaneContent
        initialNodeList={initialNodeList}
        initialCategoryList={initialCategoryList}
        initialChannelList={initialChannelList}
      />
    );
  } catch (e) {
    return <NavigationPaneContent />;
  }
}

type Props = {
  initialNodeList?: NodeListResult;
  initialCategoryList?: CategoryListOKResponse;
  initialChannelList?: ChannelListOKResponse;
};

function NavigationPaneContent({
  initialNodeList,
  initialCategoryList,
  initialChannelList,
}: Props) {
  return (
    <styled.header
      display="flex"
      height="full"
      alignItems="end"
      flexDirection="column"
      borderRadius="md"
      className={Floating()}
    >
      <AdminZone />
      <Box id="desktop-nav-box" w="full" height="full" minH="0" p="2">
        <ContentNavigationList
          initialNodeList={initialNodeList}
          initialCategoryList={initialCategoryList}
          initialChannelList={initialChannelList}
        />
      </Box>
    </styled.header>
  );
}
