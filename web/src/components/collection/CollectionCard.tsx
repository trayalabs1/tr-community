import Link from "next/link";
import { memo } from "react";

import { Collection } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { styled } from "@/styled-system/jsx";

import { CollectionIcon } from "../ui/icons/Collection";
import { CollectionMenu } from "./CollectionMenu/CollectionMenu";
import { ProfileHoverTooltip } from "../post/ProfileHoverTooltip";
import { TRAYA_COLORS } from "@/theme/traya-colors";

type Props = {
  collection: Collection;
  hideOwner?: boolean;
};

export const CollectionCard = memo(({ collection, hideOwner = false }: Props) => {
  const session = useSession();
  const url = `/c/${collection.slug}`;
  const itemsLabel = collection.item_count === 1 ? "item" : "items";

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: `1px solid ${TRAYA_COLORS.border.default}`,
        boxShadow: TRAYA_COLORS.shadow.subtle,
        transition: "all 0.2s ease-in-out",
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = TRAYA_COLORS.shadow.medium;
      }}
      onMouseLeave={(e) => {
        const el = e.currentTarget as HTMLElement;
        el.style.boxShadow = TRAYA_COLORS.shadow.subtle;
      }}
    >
      {/* Header Section - Owner Info and Menu */}
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
          {!hideOwner && (
            <ProfileHoverTooltip profile={collection.owner}>
              <Link href={`/m/${collection.owner.handle}`} style={{ textDecoration: "none" }}>
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
                    background: TRAYA_COLORS.gradient,
                    color: "white",
                    border: "none",
                    cursor: "pointer",
                    transition: "opacity 0.2s ease-in-out",
                    flexShrink: 0,
                  }}
                >
                  {collection.owner.handle.charAt(0).toUpperCase()}
                </styled.button>
              </Link>
            </ProfileHoverTooltip>
          )}

          <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
            <Link href={`/m/${collection.owner.handle}`} style={{ textDecoration: "none" }}>
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
                @{collection.owner.handle}
              </styled.p>
            </Link>
            <styled.div display="flex" alignItems="center" gap="1.5" style={{ margin: "0" }}>
              <CollectionIcon width="4" height="4" style={{ color: "var(--colors-fg-muted)" }} />
              <styled.p
                fontSize="xs"
                color="fg.muted"
                style={{
                  margin: "0",
                }}
              >
                {collection.item_count} {itemsLabel}
              </styled.p>
            </styled.div>
          </styled.div>
        </styled.div>

        {session && (
          <styled.div
            style={{
              opacity: "0.7",
              transition: "opacity 0.2s ease-in-out",
              flexShrink: 0,
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = "1";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = "0.7";
            }}
          >
            <CollectionMenu collection={collection} />
          </styled.div>
        )}
      </styled.div>

      {/* Description Section */}
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
            {collection.description || collection.name}
          </styled.p>
        </Link>
      </styled.div>
    </styled.div>
  );
});

CollectionCard.displayName = "CollectionCard";
