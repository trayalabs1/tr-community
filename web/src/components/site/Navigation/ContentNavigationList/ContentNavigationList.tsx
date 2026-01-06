"use client";

import { useState, useEffect } from "react";
import { usePathname } from "next/navigation";
import {
  CategoryListOKResponse,
  ChannelListOKResponse,
} from "@/api/openapi-schema";
import { useChannelList, useChannelCategoryList } from "@/api/openapi-client/channels";
import { useNodeList } from "@/api/openapi-client/nodes";
import { ChannelList } from "@/components/channel/ChannelList/ChannelList";
import { TopicsSection } from "./TopicsSection";
import { LStack, styled } from "@/styled-system/jsx";
import { LibraryIcon } from "@/components/ui/icons/Library";
import { LibraryNavigationTree } from "@/components/site/Navigation/LibraryNavigationTree/LibraryNavigationTree";
import { css } from "@/styled-system/css";

import { CollectionsAnchor } from "../Anchors/Collections";
import { LinksAnchor } from "../Anchors/Link";
import { MembersAnchor } from "../Anchors/Members";

type Props = {
  initialCategoryList?: CategoryListOKResponse;
  initialChannelList?: ChannelListOKResponse;
};

export function ContentNavigationList(props: Props) {
  const pathname = usePathname();

  const { data: channelListData } = useChannelList({
    swr: {
      fallbackData: props.initialChannelList,
    },
  });

  const firstChannelId = channelListData?.channels?.[0]?.id;
  const [selectedChannelId, setSelectedChannelId] = useState<string | undefined>(firstChannelId);

  useEffect(() => {
    if (firstChannelId && !selectedChannelId) {
      setSelectedChannelId(firstChannelId);
    }
  }, [firstChannelId, selectedChannelId]);

  const { data: selectedChannelCategories } = useChannelCategoryList(selectedChannelId ?? "", {
    swr: {
      enabled: !!selectedChannelId,
    },
  });

  const { data: nodeListData } = useNodeList(
    {
      visibility: ["draft", "review", "unlisted", "published"],
    },
    {
      swr: {
        enabled: true,
      },
    }
  );

  // Extract current category slug from URL: /channels/[id]/categories/[slug]
  const currentCategorySlug = pathname.includes("/categories/")
    ? pathname.split("/categories/")[1]?.split("/")[0]
    : undefined;

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
          selectedChannelID={selectedChannelId}
          onChannelSelect={setSelectedChannelId}
        />
        <div
          className={css({
            mt: "4",
            mb: "4",
            bg: "border.default",
          })}
          style={{
            width: "calc(100% + 2rem)",
            marginLeft: "calc(-1rem)",
            height: "0.5px",
          }}
        />
        {selectedChannelId && selectedChannelCategories && (
          <TopicsSection
            categories={selectedChannelCategories}
            channelID={selectedChannelId}
            currentCategorySlug={currentCategorySlug}
          />
        )}
        {nodeListData?.nodes && nodeListData.nodes.length > 0 && (
          <LStack gap="1">
            <div
              className={css({
                display: "flex",
                alignItems: "center",
                gap: "1",
                py: "1",
                fontSize: "xs",
                fontWeight: "semibold",
                color: "fg.muted",
                textTransform: "uppercase",
                letterSpacing: "wider",
              })}
            >
              <LibraryIcon width="3" height="3" />
              <span>Library</span>
            </div>
            <div
              className={css({
                ps: "4",
              })}
            >
              <LibraryNavigationTree
                initialNodeList={nodeListData}
                currentNode={undefined}
                visibility={["draft", "review", "unlisted", "published"]}
                hideHeader={true}
              />
            </div>
          </LStack>
        )}
      </LStack>

      <LStack gap="1">
        <CollectionsAnchor />
        <LinksAnchor />
        <MembersAnchor />
      </LStack>
    </styled.nav>
  );
}
