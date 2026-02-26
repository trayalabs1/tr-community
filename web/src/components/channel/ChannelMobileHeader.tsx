"use client";

import Link from "next/link";
import { Account, Channel } from "@/api/openapi-schema";
import { MembersIcon } from "@/components/ui/icons/Members";
import { ProfileIcon } from "@/components/ui/icons/Profile";
import { HeaderWithBackArrow } from "@/components/site/Header";
import { HStack, VStack, styled } from "@/styled-system/jsx";
import { getAssetURL } from "@/utils/asset";
import { ChannelFilterBar } from "@/components/channel/ChannelFilterBar";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { BookmarkButton } from "@/components/channel/BookmarkButton";
import { NotificationButton } from "@/components/channel/NotificationButton";
import { useChannelPermissions } from "@/lib/channel/permissions";

type ChannelMobileHeaderProps = {
  channel: Channel;
  session?: Account;
  categories: any[];
  selectedCategorySlug: string | null;
  selectedVisibility: string | null;
  onCategoryChange: (slug: string | null) => void;
  onVisibilityChange: (visibility: string | null) => void;
  hasUnreadNotifications?: boolean;
  bookmarkCount?: number;
};

export function ChannelMobileHeader({
  channel,
  session,
  categories,
  selectedCategorySlug,
  selectedVisibility,
  onCategoryChange,
  onVisibilityChange,
  hasUnreadNotifications = false,
  bookmarkCount = 0,
}: ChannelMobileHeaderProps) {
  const permissions = useChannelPermissions(channel.id);
  const isAdmin = permissions.isOwner || permissions.isAdmin;

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
      {isAdmin ? (
        <HeaderWithBackArrow
          title={channel.name}
          subtitle={channel.description || undefined}
          headerIcon={channelIcon}
          headerIconBackground="transparent"
          mobileOnly
          showBorder={false}
        />
      ) : (
        <HStack
          display={{ base: "flex", md: "none" }}
          alignItems="center"
          gap="3"
          p="4"
          borderBottomWidth="thin"
          borderBottomColor="border.default"
          justifyContent="space-between"
          bg="white"
          width="full"
          style={{
            borderBottomColor: TRAYA_COLORS.neutral.border,
          }}
        >
          <HStack gap="3" flex="1" alignItems="center">
            {session ? (
              <Link href={`/m/${session.handle}`} style={{ textDecoration: "none" }}>
                <styled.div
                  w="8"
                  h="8"
                  rounded="full"
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  flexShrink="0"
                  style={{
                    background: TRAYA_COLORS.gradient,
                  }}
                >
                  <ProfileIcon style={{ color: "white", width: "18px", height: "18px" }} />
                </styled.div>
              </Link>
            ) : (
              channelIcon
            )}
            <VStack alignItems="start" gap="0.5">
              <styled.h3
                fontWeight="semibold"
                color="fg.default"
                style={{
                  margin: "0",
                  fontSize: "16px",
                }}
              >
                {channel.name}
              </styled.h3>
              {channel.description && (
                <styled.p
                  fontSize="xs"
                  color="fg.muted"
                  style={{
                    margin: "0",
                    lineHeight: "1.3",
                  }}
                  lineClamp={1}
                >
                  {channel.description}
                </styled.p>
              )}
            </VStack>
          </HStack>
          <HStack gap="2" flexShrink="0">
            <BookmarkButton count={bookmarkCount} />
            <NotificationButton hasUnread={hasUnreadNotifications} />
          </HStack>
        </HStack>
      )}

      {/* Filter and Create Post Bar */}
      <styled.div display={{ base: "block", md: "none" }} width="full" px="4" pt="3" pb="0">
        <ChannelFilterBar
          channelID={channel?.id}
          categories={categories}
          selectedCategorySlug={selectedCategorySlug}
          selectedVisibility={selectedVisibility}
          onCategoryChange={onCategoryChange}
          onVisibilityChange={onVisibilityChange}
        />
      </styled.div>
    </VStack>
  );
}
