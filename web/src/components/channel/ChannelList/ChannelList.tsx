"use client";

import { useChannelList } from "@/api/openapi-client/channels";
import {
  Channel,
  ChannelListOKResponse,
  CategoryListOKResponse,
  NodeListResult,
} from "@/api/openapi-schema";
import { Anchor } from "@/components/site/Anchor";
import { NavigationHeader } from "@/components/site/Navigation/ContentNavigationList/NavigationHeader";
import { LibraryNavigationTree } from "@/components/site/Navigation/LibraryNavigationTree/LibraryNavigationTree";
import { Unready } from "@/components/site/Unready";
import { BulletIcon } from "@/components/ui/icons/Bullet";
import { CategoryIcon } from "@/components/ui/icons/Category";
import { DiscussionIcon } from "@/components/ui/icons/Discussion";
import { LibraryIcon } from "@/components/ui/icons/Library";
import { css } from "@/styled-system/css";
import { HStack, LStack } from "@/styled-system/jsx";
import { CategoryListTree } from "@/components/category/CategoryList/CategoryList";

export type Props = {
  initialChannelList?: ChannelListOKResponse;
  currentChannelID?: string;
  channelCategories?: CategoryListOKResponse;
  channelNodes?: NodeListResult;
};

export function ChannelList({
  initialChannelList,
  currentChannelID,
  channelCategories,
  channelNodes,
}: Props) {
  const { data, error } = useChannelList({
    swr: {
      fallbackData: initialChannelList,
    },
  });

  if (!data) {
    return <Unready error={error} />;
  }

  return (
    <LStack gap="0">
      <NavigationHeader href="/channels">
        <HStack gap="1">
          <CategoryIcon />
          Channels
        </HStack>
      </NavigationHeader>

      {data.channels.length > 0 ? (
        <LStack gap="1">
          {data.channels.map((channel) => (
            <ChannelListItem
              key={channel.id}
              channel={channel}
              isHighlighted={channel.id === currentChannelID}
              isExpanded={channel.id === currentChannelID}
              categories={
                channel.id === currentChannelID ? channelCategories : undefined
              }
              nodes={channel.id === currentChannelID ? channelNodes : undefined}
            />
          ))}
        </LStack>
      ) : null}
    </LStack>
  );
}

type ChannelListItemProps = {
  channel: Channel;
  isHighlighted: boolean;
  isExpanded: boolean;
  categories?: CategoryListOKResponse;
  nodes?: NodeListResult;
};

function ChannelListItem({
  channel,
  isHighlighted,
  isExpanded,
  categories,
  nodes,
}: ChannelListItemProps) {
  return (
    <LStack gap="1">
      <Anchor
        href={`/channels/${channel.id}`}
        className={css({
          display: "flex",
          alignItems: "center",
          gap: "1",
          ps: "6",
          py: "1",
          pr: "1",
          h: "8",
          fontSize: "sm",
          fontWeight: "medium",
          borderRadius: "l2",
          color: "fg.subtle",
          background: isHighlighted ? "bg.selected" : undefined,
          _hover: {
            textDecoration: "none",
            background: "bg.emphasized",
            color: "fg.emphasized",
          },
        })}
      >
        <BulletIcon width="4" height="4" />
        <span
          className={css({
            fontSize: "xs",
            width: "full",
            textWrap: "nowrap",
            textOverflow: "ellipsis",
            overflowX: "hidden",
          })}
        >
          {channel.name}
        </span>
      </Anchor>

      {isExpanded && (
        <div
          className={css({
            ps: "8",
            ms: "6",
            py: "1",
          })}
          style={{
            borderLeft: "2px solid var(--colors-border-subtle)",
          }}
        >
          <ChannelContentNavigation
            channelID={channel.id}
            categories={categories}
            nodes={nodes}
          />
        </div>
      )}
    </LStack>
  );
}

type ChannelContentNavigationProps = {
  channelID: string;
  categories?: CategoryListOKResponse;
  nodes?: NodeListResult;
};

function ChannelContentNavigation({
  channelID,
  categories,
  nodes,
}: ChannelContentNavigationProps) {
  return (
    <LStack gap="2">
      {categories && categories.categories.length > 0 && (
        <LStack gap="1">
          <div
            className={css({
              display: "flex",
              alignItems: "center",
              gap: "1",
              ps: "2",
              py: "1",
              fontSize: "xs",
              fontWeight: "semibold",
              color: "fg.muted",
              textTransform: "uppercase",
              letterSpacing: "wider",
            })}
          >
            <DiscussionIcon width="3" height="3" />
            <span>Categories</span>
          </div>
          <CategoryListTree
            categories={categories.categories}
            channelID={channelID}
            mutate={async () => undefined}
            hideHeader={true}
          />
        </LStack>
      )}
      {nodes && nodes.nodes.length > 0 && (
        <LStack gap="1">
          <div
            className={css({
              display: "flex",
              alignItems: "center",
              gap: "1",
              ps: "2",
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
          <LibraryNavigationTree
            initialNodeList={nodes}
            currentNode={undefined}
            visibility={["draft", "review", "unlisted", "published"]}
            hideHeader={true}
          />
        </LStack>
      )}
    </LStack>
  );
}
