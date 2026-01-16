"use client";

import { useRouter } from "next/navigation";
import { ArrowLeft, Info, Users } from "lucide-react";

import { Channel } from "@/api/openapi-schema";
import { MembersIcon } from "@/components/ui/icons/Members";
import { Heading } from "@/components/ui/heading";
import { HStack, VStack, styled } from "@/styled-system/jsx";
import { getAssetURL } from "@/utils/asset";
import { ChannelFilterBar } from "@/components/channel/ChannelFilterBar";

type ChannelMobileHeaderProps = {
  channel: Channel;
  categories: any[];
  selectedCategorySlug: string | null;
  selectedVisibility: string | null;
  onCategoryChange: (slug: string | null) => void;
  onVisibilityChange: (visibility: string | null) => void;
};

export function ChannelMobileHeader({
  channel,
  categories,
  selectedCategorySlug,
  selectedVisibility,
  onCategoryChange,
  onVisibilityChange
}: ChannelMobileHeaderProps) {
  const router = useRouter();

  return (
    <VStack alignItems="start" gap="0" width="full">
      {/* Header with back, channel info, and action buttons */}
      <HStack
        display={{ base: "flex", md: "none" }}
        alignItems="center"
        gap="3"
        width="full"
        p="4"
      >
        <styled.button
          onClick={() => router.back()}
          display="flex"
          alignItems="center"
          justifyContent="center"
          flexShrink="0"
          style={{
            backgroundColor: "transparent",
            border: "none",
            cursor: "pointer",
            padding: "0",
            color: "var(--colors-fg-default)",
          }}
        >
          <ArrowLeft size={20} strokeWidth={2} />
        </styled.button>

        {channel?.icon ? (
          <styled.img
            src={getAssetURL(channel.icon.path)}
            alt={channel.name}
            w="10"
            h="10"
            rounded="lg"
            objectFit="cover"
            flexShrink="0"
          />
        ) : (
          <styled.div
            w="10"
            h="10"
            rounded="lg"
            display="flex"
            alignItems="center"
            justifyContent="center"
            flexShrink="0"
            style={{
              background: "#4a9d6f",
            }}
          >
            <MembersIcon width="5" height="5" style={{ color: "#ffffff" }} />
          </styled.div>
        )}

        <VStack alignItems="start" gap="0.5" width="full" minWidth="0">
          <Heading as="h2" size="md" style={{ overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", width: "100%" }}>
            {channel.name}
          </Heading>
          {channel.description && (
            <styled.p
              fontSize="xs"
              color="fg.muted"
              style={{
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
                width: "100%",
                margin: "0",
              }}
            >
              {channel.description}
            </styled.p>
          )}
        </VStack>

        {/* <HStack gap="2" alignItems="center" flexShrink="0">
          <styled.button
            display="flex"
            alignItems="center"
            justifyContent="center"
            style={{
              width: "36px",
              height: "36px",
              borderRadius: "50%",
              backgroundColor: "transparent",
              border: "1px solid var(--colors-border-default)",
              cursor: "pointer",
              color: "var(--colors-fg-default)",
              padding: "0",
            }}
          >
            <Info size={18} strokeWidth={2} />
          </styled.button>
          <styled.button
            display="flex"
            alignItems="center"
            justifyContent="center"
            style={{
              width: "36px",
              height: "36px",
              borderRadius: "50%",
              backgroundColor: "transparent",
              border: "1px solid var(--colors-border-default)",
              cursor: "pointer",
              color: "var(--colors-fg-default)",
              padding: "0",
            }}
          >
            <Users size={18} strokeWidth={2} />
          </styled.button>
        </HStack> */}
      </HStack>

      {/* Filter and Create Post Bar */}
      <styled.div display={{ base: "block", md: "none" }} width="full" px="4" py="3">
        <ChannelFilterBar
          channelID={channel?.id}
          categories={categories}
          selectedCategorySlug={selectedCategorySlug}
          selectedVisibility={selectedVisibility}
          onCategoryChange={onCategoryChange}
          onVisibilityChange={onVisibilityChange}
        />
      </styled.div>

      {/* Bottom divider */}
      <styled.div
        width="full"
        style={{
          height: "1px",
          backgroundColor: "var(--colors-border-default)",
          marginLeft: "-100vw",
          marginRight: "-100vw",
          paddingLeft: "100vw",
          paddingRight: "100vw",
        }}
      />
    </VStack>
  );
}
