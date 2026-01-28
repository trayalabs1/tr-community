"use client";

import { useState } from "react";
import Link from "next/link";
import { useChannelList } from "@/api/openapi-client/channels";
import { ChannelListOKResponse } from "@/api/openapi-schema";
import { Unready } from "@/components/site/Unready";
import { CategoryIcon } from "@/components/ui/icons/Category";
import { ChevronDownIcon } from "@/components/ui/icons/Chevron";
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
  const [isExpanded, setIsExpanded] = useState(true);
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
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className={css({
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "1",
          py: "1",
          pr: "1",
          h: "8",
          fontSize: "xs",
          fontWeight: "semibold",
          color: "fg.muted",
          textTransform: "uppercase",
          letterSpacing: "wider",
          background: "transparent",
          border: "none",
          cursor: "pointer",
          w: "full",
          textAlign: "left",
          _hover: {
            color: "fg.subtle",
          },
        })}
      >
        <HStack gap="1">
          <CategoryIcon width="4" height="4" />
          <span>Channels</span>
        </HStack>
        <ChevronDownIcon
          width="4"
          height="4"
          style={{
            transition: "transform 200ms",
            transform: isExpanded ? "rotate(0deg)" : "rotate(-90deg)",
          }}
        />
      </button>

      {isExpanded && data.channels.length > 0 ? (
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
