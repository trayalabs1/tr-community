import Link from "next/link";
import chroma from "chroma-js";

import {
  DatagraphItem,
  DatagraphItemKind,
  DatagraphItemNode,
  DatagraphItemPost,
  DatagraphItemProfile,
  DatagraphItemReply,
  DatagraphItemThread,
} from "@/api/openapi-schema";
import { styled } from "@/styled-system/jsx";
import { ColorPalette } from "@/styled-system/tokens";

import { Timestamp } from "../site/Timestamp";
import { Badge } from "../ui/badge";
import { ProfileHoverTooltip } from "../post/ProfileHoverTooltip";

type Props = {
  item: DatagraphItem;
};

export function DatagraphItemCard({ item }: Props) {
  switch (item.kind) {
    case DatagraphItemKind.post:
      return <DatagraphItemPostGenericCard item={item} />;

    case DatagraphItemKind.thread:
      return <DatagraphItemPostGenericCard item={item} />;

    case DatagraphItemKind.reply:
      return <DatagraphItemReplyCard item={item} />;

    case DatagraphItemKind.node:
      return <DatagraphItemNodeCard item={item} />;

    // case DatagraphItemKind.collection:
    //   return null;

    case DatagraphItemKind.profile:
      return <DatagraphItemProfileCard item={item} />;

    // case DatagraphItemKind.event:
    //   return null;
  }
}

export function DatagraphItemPostGenericCard({
  item,
}: {
  item: DatagraphItemPost | DatagraphItemThread;
}) {
  const { ref } = item;
  const url = `/t/${ref.slug}`;

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: "1px solid #f1f2f4",
        boxShadow: "0 0.5px 2px rgba(0, 0, 0, 0.08)",
        transition: "all 0.2s ease-in-out",
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.12)";
      }}
      onMouseLeave={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 0.5px 2px rgba(0, 0, 0, 0.08)";
      }}
    >
      {/* Header Section - Author Info */}
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        gap="3"
        px="4"
        pt="4"
        pb="2"
      >
        <styled.div display="flex" alignItems="center" gap="3" flex="1" minW="0">
          <ProfileHoverTooltip profile={ref.author}>
            <Link href={`/m/${ref.author.handle}`} style={{ textDecoration: "none" }}>
              <styled.button
                w="10"
                h="10"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                fontSize="sm"
                fontWeight="semibold"
                style={{
                  backgroundColor: "#f0f5f1",
                  color: "#2D7A4A",
                  border: "none",
                  cursor: "pointer",
                  transition: "opacity 0.2s ease-in-out",
                  flexShrink: 0,
                }}
              >
                {ref.author.name.charAt(0).toUpperCase()}
              </styled.button>
            </Link>
          </ProfileHoverTooltip>

          <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
            <Link href={`/m/${ref.author.handle}`} style={{ textDecoration: "none" }}>
              <styled.p
                fontSize="sm"
                fontWeight="semibold"
                color="fg.default"
                style={{
                  margin: "0",
                  transition: "color 0.2s ease-in-out",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  cursor: "pointer",
                }}
              >
                {ref.author.name}
              </styled.p>
            </Link>
            <styled.p
              fontSize="xs"
              color="fg.muted"
              style={{
                margin: "0",
              }}
            >
              <Timestamp created={ref.createdAt} />
            </styled.p>
          </styled.div>
        </styled.div>
      </styled.div>

      {/* Content Section - Description */}
      <styled.div
        display="flex"
        flexDir="column"
        gap="1"
        px="4"
        pt="2"
        pb="4"
      >
        <Link href={url} style={{ textDecoration: "none" }}>
          <styled.p
            fontSize="sm"
            color="fg.default"
            fontWeight="medium"
            style={{
              margin: "0",
              lineHeight: "1.6",
              whiteSpace: "pre-wrap",
              overflow: "hidden",
              display: "-webkit-box",
              WebkitLineClamp: 3,
              WebkitBoxOrient: "vertical",
              textOverflow: "ellipsis",
              cursor: "pointer",
            }}
          >
            {(ref.description ?? ref.title) || "(untitled post)"}
          </styled.p>
        </Link>

        <styled.div display="flex" gap="2.5" flexWrap="wrap">
          <DatagraphItemBadge kind={item.kind} />
        </styled.div>
      </styled.div>
    </styled.div>
  );
}

export function DatagraphItemReplyCard({ item }: { item: DatagraphItemReply }) {
  const { ref } = item;
  const url = `/t/locate/${ref.id}`;

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: "1px solid #f1f2f4",
        boxShadow: "0 0.5px 2px rgba(0, 0, 0, 0.08)",
        transition: "all 0.2s ease-in-out",
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.12)";
      }}
      onMouseLeave={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 0.5px 2px rgba(0, 0, 0, 0.08)";
      }}
    >
      {/* Header Section - Author Info */}
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        gap="3"
        px="4"
        pt="4"
        pb="2"
      >
        <styled.div display="flex" alignItems="center" gap="3" flex="1" minW="0">
          <ProfileHoverTooltip profile={ref.author}>
            <Link href={`/m/${ref.author.handle}`} style={{ textDecoration: "none" }}>
              <styled.button
                w="10"
                h="10"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                fontSize="sm"
                fontWeight="semibold"
                style={{
                  backgroundColor: "#f0f5f1",
                  color: "#2D7A4A",
                  border: "none",
                  cursor: "pointer",
                  transition: "opacity 0.2s ease-in-out",
                  flexShrink: 0,
                }}
              >
                {ref.author.name.charAt(0).toUpperCase()}
              </styled.button>
            </Link>
          </ProfileHoverTooltip>

          <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
            <Link href={`/m/${ref.author.handle}`} style={{ textDecoration: "none" }}>
              <styled.p
                fontSize="sm"
                fontWeight="semibold"
                color="fg.default"
                style={{
                  margin: "0",
                  transition: "color 0.2s ease-in-out",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  cursor: "pointer",
                }}
              >
                {ref.author.name}
              </styled.p>
            </Link>
            <styled.p
              fontSize="xs"
              color="fg.muted"
              style={{
                margin: "0",
              }}
            >
              <Timestamp created={ref.createdAt} />
            </styled.p>
          </styled.div>
        </styled.div>
      </styled.div>

      {/* Content Section - Description */}
      <styled.div
        display="flex"
        flexDir="column"
        gap="1"
        px="4"
        pt="2"
        pb="4"
      >
        <Link href={url} style={{ textDecoration: "none" }}>
          <styled.p
            fontSize="sm"
            color="fg.default"
            fontWeight="medium"
            style={{
              margin: "0",
              lineHeight: "1.6",
              whiteSpace: "pre-wrap",
              overflow: "hidden",
              display: "-webkit-box",
              WebkitLineClamp: 3,
              WebkitBoxOrient: "vertical",
              textOverflow: "ellipsis",
              cursor: "pointer",
            }}
          >
            {(ref.description ?? ref.title) || "(untitled thread)"}
          </styled.p>
        </Link>

        <styled.div display="flex" gap="2.5" flexWrap="wrap">
          <DatagraphItemBadge kind={item.kind} />
        </styled.div>
      </styled.div>
    </styled.div>
  );
}

export function DatagraphItemNodeCard({ item }: { item: DatagraphItemNode }) {
  const { ref } = item;
  const url = `/l/${ref.slug}`;

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: "1px solid #f1f2f4",
        boxShadow: "0 0.5px 2px rgba(0, 0, 0, 0.08)",
        transition: "all 0.2s ease-in-out",
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.12)";
      }}
      onMouseLeave={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 0.5px 2px rgba(0, 0, 0, 0.08)";
      }}
    >
      {/* Header Section - Owner Info */}
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        gap="3"
        px="4"
        pt="4"
        pb="2"
      >
        <styled.div display="flex" alignItems="center" gap="3" flex="1" minW="0">
          <ProfileHoverTooltip profile={ref.owner}>
            <Link href={`/m/${ref.owner.handle}`} style={{ textDecoration: "none" }}>
              <styled.button
                w="10"
                h="10"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                fontSize="sm"
                fontWeight="semibold"
                style={{
                  backgroundColor: "#f0f5f1",
                  color: "#2D7A4A",
                  border: "none",
                  cursor: "pointer",
                  transition: "opacity 0.2s ease-in-out",
                  flexShrink: 0,
                }}
              >
                {ref.owner.name.charAt(0).toUpperCase()}
              </styled.button>
            </Link>
          </ProfileHoverTooltip>

          <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
            <Link href={`/m/${ref.owner.handle}`} style={{ textDecoration: "none" }}>
              <styled.p
                fontSize="sm"
                fontWeight="semibold"
                color="fg.default"
                style={{
                  margin: "0",
                  transition: "color 0.2s ease-in-out",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  cursor: "pointer",
                }}
              >
                {ref.owner.name}
              </styled.p>
            </Link>
            <styled.p
              fontSize="xs"
              color="fg.muted"
              style={{
                margin: "0",
              }}
            >
              <Timestamp created={ref.createdAt} />
            </styled.p>
          </styled.div>
        </styled.div>
      </styled.div>

      {/* Content Section - Description */}
      <styled.div
        display="flex"
        flexDir="column"
        gap="1"
        px="4"
        pt="2"
        pb="4"
      >
        <Link href={url} style={{ textDecoration: "none" }}>
          <styled.p
            fontSize="sm"
            color="fg.default"
            fontWeight="medium"
            style={{
              margin: "0",
              lineHeight: "1.6",
              whiteSpace: "pre-wrap",
              overflow: "hidden",
              display: "-webkit-box",
              WebkitLineClamp: 3,
              WebkitBoxOrient: "vertical",
              textOverflow: "ellipsis",
              cursor: "pointer",
            }}
          >
            {ref.description || ref.name}
          </styled.p>
        </Link>

        <styled.div display="flex" gap="2.5" flexWrap="wrap">
          <DatagraphItemBadge kind={item.kind} />
        </styled.div>
      </styled.div>
    </styled.div>
  );
}

export function DatagraphItemProfileCard({
  item,
}: {
  item: DatagraphItemProfile;
}) {
  const { ref } = item;
  const url = `/m/${ref.handle}`;

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: "1px solid #f1f2f4",
        boxShadow: "0 0.5px 2px rgba(0, 0, 0, 0.08)",
        transition: "all 0.2s ease-in-out",
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 4px 12px rgba(0, 0, 0, 0.12)";
      }}
      onMouseLeave={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = "0 0.5px 2px rgba(0, 0, 0, 0.08)";
      }}
    >
      {/* Header Section - Profile Info */}
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        gap="3"
        px="4"
        pt="4"
        pb="2"
      >
        <styled.div display="flex" alignItems="center" gap="3" flex="1" minW="0">
          <ProfileHoverTooltip profile={ref}>
            <Link href={url} style={{ textDecoration: "none" }}>
              <styled.button
                w="10"
                h="10"
                rounded="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                fontSize="sm"
                fontWeight="semibold"
                style={{
                  backgroundColor: "#f0f5f1",
                  color: "#2D7A4A",
                  border: "none",
                  cursor: "pointer",
                  transition: "opacity 0.2s ease-in-out",
                  flexShrink: 0,
                }}
              >
                {ref.name.charAt(0).toUpperCase()}
              </styled.button>
            </Link>
          </ProfileHoverTooltip>

          <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
            <Link href={url} style={{ textDecoration: "none" }}>
              <styled.p
                fontSize="sm"
                fontWeight="semibold"
                color="fg.default"
                style={{
                  margin: "0",
                  transition: "color 0.2s ease-in-out",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  cursor: "pointer",
                }}
              >
                {ref.name}
              </styled.p>
            </Link>
            <styled.p
              fontSize="xs"
              color="fg.muted"
              style={{
                margin: "0",
              }}
            >
              @{ref.handle}
            </styled.p>
          </styled.div>
        </styled.div>
      </styled.div>

      {/* Content Section - Bio */}
      <styled.div
        display="flex"
        flexDir="column"
        gap="1"
        px="4"
        pt="2"
        pb="4"
      >
        {ref.bio && (
          <Link href={url} style={{ textDecoration: "none" }}>
            <styled.p
              fontSize="sm"
              color="fg.default"
              fontWeight="medium"
              style={{
                margin: "0",
                lineHeight: "1.6",
                whiteSpace: "pre-wrap",
                overflow: "hidden",
                display: "-webkit-box",
                WebkitLineClamp: 3,
                WebkitBoxOrient: "vertical",
                textOverflow: "ellipsis",
                cursor: "pointer",
              }}
            >
              {ref.bio}
            </styled.p>
          </Link>
        )}

        <styled.div display="flex" gap="2.5" flexWrap="wrap">
          <DatagraphItemBadge kind={item.kind} />
        </styled.div>
      </styled.div>
    </styled.div>
  );
}

export function DatagraphItemBadge({ kind }: { kind: DatagraphItemKind }) {
  const label = getDatagraphKindLabel(kind);
  const colour = getDatagraphKindColour(kind);

  const cssVars = badgeColourCSS(colour);

  return (
    <Badge
      style={cssVars}
      backgroundColor="var(--colors-color-palette-bg)"
      borderColor="var(--colors-color-palette-bo)"
      color="var(--colors-color-palette-fg)"
    >
      {label}
    </Badge>
  );
}

export function badgeColourCSS(c: string) {
  const { bg, bo, fg } = badgeColours(c);

  return {
    "--colors-color-palette-fg": fg,
    "--colors-color-palette-bo": bo,
    "--colors-color-palette-bg": bg,
  } as React.CSSProperties;
}

export function badgeColours(c: string) {
  const colour = chroma(c);

  const bg = colour.luminance(0.8).css();
  const bo = colour.luminance(0.6).saturate(1.3).css();
  const fg = colour.darken(1.5).saturate(2).css();

  return { bg, bo, fg };
}

function getDatagraphKindLabel(kind: DatagraphItemKind): string {
  switch (kind) {
    case DatagraphItemKind.post:
      return "Post";
    case DatagraphItemKind.thread:
      return "Thread";
    case DatagraphItemKind.reply:
      return "Reply";
    case DatagraphItemKind.node:
      return "Library";
    case DatagraphItemKind.collection:
      return "Collection";
    case DatagraphItemKind.profile:
      return "Profile";
    case DatagraphItemKind.event:
      return "Event";
  }
}

function getDatagraphKindColour(kind: DatagraphItemKind): ColorPalette {
  switch (kind) {
    case DatagraphItemKind.post:
      return "pink";
    case DatagraphItemKind.thread:
      return "pink";
    case DatagraphItemKind.reply:
      return "pink";
    case DatagraphItemKind.node:
      return "green";
    case DatagraphItemKind.collection:
      return "blue";
    case DatagraphItemKind.profile:
      return "red";
    case DatagraphItemKind.event:
      return "amber";
  }
}
