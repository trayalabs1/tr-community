"use client";

import { Channel } from "@/api/openapi-schema";
import { MembersIcon } from "@/components/ui/icons/Members";
import { HeaderWithBackArrow } from "@/components/site/Header";
import { VStack, styled } from "@/styled-system/jsx";
import { getAssetURL } from "@/utils/asset";
import { ChannelFilterBar } from "@/components/channel/ChannelFilterBar";
import { TRAYA_COLORS } from "@/theme/traya-colors";

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

  const channelIcon = channel?.icon ? (
    <styled.img
      src={getAssetURL(channel.icon.path)}
      alt={channel.name}
      w="8"
      h="8"
      rounded="lg"
      objectFit="cover"
      flexShrink="0"
    />
  ) : (
    <styled.div
      w="8"
      h="8"
      rounded="lg"
      display="flex"
      alignItems="center"
      justifyContent="center"
      flexShrink="0"
      style={{
        background: TRAYA_COLORS.secondary,
      }}
    >
      <MembersIcon width="5" height="5" style={{ color: TRAYA_COLORS.primary }} />
    </styled.div>
  );

  return (
    <VStack alignItems="start" gap="0" width="full">
      <HeaderWithBackArrow
        title={channel.name}
        subtitle={channel.description || undefined}
        headerIcon={channelIcon}
        headerIconBackground="transparent"
        mobileOnly
        showBorder={false}
      />

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
        }}
      />
    </VStack>
  );
}
