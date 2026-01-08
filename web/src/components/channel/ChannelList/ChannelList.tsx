"use client";

import Link from "next/link";
import { useChannelList } from "@/api/openapi-client/channels";
import { ChannelListOKResponse } from "@/api/openapi-schema";
import { NavigationHeader } from "@/components/site/Navigation/ContentNavigationList/NavigationHeader";
import { Unready } from "@/components/site/Unready";
import { CategoryIcon } from "@/components/ui/icons/Category";
import { css } from "@/styled-system/css";
import { HStack, LStack } from "@/styled-system/jsx";

export type Props = {
  initialChannelList?: ChannelListOKResponse;
  selectedChannelID?: string;
  onChannelSelect?: (channelId: string) => void;
};

export function ChannelList({
  initialChannelList,
  selectedChannelID,
  onChannelSelect,
}: Props) {
  const { data, error } = useChannelList({
    swr: {
      fallbackData: initialChannelList,
    },
  });

  if (!data) {
    const isUnauthorized = error && "status" in error && error.status === 401;
    if (isUnauthorized) {
      return null;
    }
    return <Unready error={error} />;
  }

  return (
    <LStack gap="0">
      <NavigationHeader href="/channels">
        <HStack gap="1">
          <CategoryIcon />
          <span
            className={css({
              textTransform: "uppercase",
              fontSize: "xs",
              fontWeight: "semibold",
              letterSpacing: "wider",
            })}
          >
            Channels
          </span>
        </HStack>
      </NavigationHeader>

      {data.channels.length > 0 ? (
        <LStack gap="1">
          {data.channels.map((channel) => (
            <Link
              key={channel.id}
              href={`/channels/${channel.id}`}
              onClick={() => onChannelSelect?.(channel.id)}
              className={css({
                display: "flex",
                alignItems: "center",
                gap: "2",
                ps: "6",
                py: "1.5",
                pr: "2",
                h: "10",
                fontSize: "sm",
                fontWeight: channel.id === selectedChannelID ? "bold" : "medium",
                borderRadius: "md",
                color: channel.id === selectedChannelID ? "fg.emphasized" : "fg.subtle",
                background: channel.id === selectedChannelID ? "bg.emphasized" : undefined,
                cursor: "pointer",
                w: "full",
                textAlign: "left",
                _hover: {
                  textDecoration: "none",
                  background: channel.id === selectedChannelID ? "bg.emphasized" : "bg.muted",
                  color: channel.id === selectedChannelID ? "fg.emphasized" : "fg.subtle",
                },
              })}
            >
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
            </Link>
          ))}
        </LStack>
      ) : null}
    </LStack>
  );
}
