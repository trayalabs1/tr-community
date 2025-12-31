"use client";

import { useChannelList } from "@/api/openapi-client/channels";
import { Channel, ChannelListOKResponse } from "@/api/openapi-schema";
import { Anchor } from "@/components/site/Anchor";
import { NavigationHeader } from "@/components/site/Navigation/ContentNavigationList/NavigationHeader";
import { Unready } from "@/components/site/Unready";
import { BulletIcon } from "@/components/ui/icons/Bullet";
import { CategoryIcon } from "@/components/ui/icons/Category";
import { css } from "@/styled-system/css";
import { HStack, LStack } from "@/styled-system/jsx";

export type Props = {
  initialChannelList?: ChannelListOKResponse;
  currentChannelID?: string;
};

export function ChannelList({
  initialChannelList,
  currentChannelID,
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
};

function ChannelListItem({ channel, isHighlighted }: ChannelListItemProps) {
  return (
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
  );
}
