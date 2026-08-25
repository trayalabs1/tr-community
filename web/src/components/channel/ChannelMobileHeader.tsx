"use client";

import Link from "next/link";
import { Account, Channel, Permission } from "@/api/openapi-schema";
import { closeWebView } from "@/lib/native/goBack";
import { useIsScrolledDown } from "@/lib/react-utility/useScrollDirection";
import { hasPermission } from "@/utils/permissions";
import { ProfileIcon } from "@/components/ui/icons/Profile";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { SearchIcon } from "@/components/ui/icons/Search";
import { NotificationIcon } from "@/components/ui/icons/Notification";
import { HStack, VStack, Box, styled } from "@/styled-system/jsx";
import { FeedFilterChips } from "@/components/channel/FeedFilterChips";
import { SelfAvatarBadge } from "@/components/member/MemberBadge/SelfAvatarBadge";
import { parsePromptNudges } from "@/components/feed/PromptNudge/prompts";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { UsersPostedToday } from "@/components/feed/QuickShare/UsersPostedToday";
import { ThreadCreateTrigger } from "@/components/thread/ThreadCreate/ThreadCreateTrigger";

type ChannelMobileHeaderProps = {
  channel: Channel;
  session?: Account;
  categories: any[];
  selectedCategorySlug: string | null;
  onCategoryChange: (slug: string | null) => void;
  selectedPrimaryTopic: string | null;
  onPrimaryTopicChange: (topic: string | null) => void;
  hasUnreadNotifications?: boolean;
};

export function ChannelMobileHeader({
  channel,
  session,
  categories,
  selectedCategorySlug,
  onCategoryChange,
  selectedPrimaryTopic,
  onPrimaryTopicChange,
  hasUnreadNotifications = false,
}: ChannelMobileHeaderProps) {
  const canManagePosts = hasPermission(session, Permission.MANAGE_POSTS);
  const isScrolledDown = useIsScrolledDown();

  return (
    <VStack
      alignItems="start"
      gap="0"
      width="full"
      position="sticky"
      top="0"
      zIndex="sticky"
      display="flex"
    >
      <VStack
        alignItems="start"
        gap="0"
        width="full"
        style={{
          background:
            "linear-gradient(180deg, rgb(235,245,240) 0%, rgb(237,246,242) 40.76%, rgb(249,252,250) 52.15%, #ffffff 100%)",
        }}
      >
        <HStack
          justifyContent="space-between"
          alignItems="center"
          width="full"
          h="14"
          pl="1"
          pr="3.5"
          style={{ backgroundColor: "#eaf5f0" }}
        >
          <styled.button
            type="button"
            onClick={closeWebView}
            display="flex"
            alignItems="center"
            justifyContent="center"
            w="9"
            h="9"
            aria-label="Back"
            style={{ background: "transparent", border: "none", cursor: "pointer", padding: "0" }}
          >
            <ArrowLeftIcon width="6" height="6" style={{ color: "var(--colors-fg-default)" }} />
          </styled.button>

          <HStack gap="2" alignItems="center" flexShrink="0">
            <Link href="/notifications" aria-label="Notifications">
              <styled.div position="relative" display="inline-flex" alignItems="center" justifyContent="center" w="10" h="10">
                <NotificationIcon width="5" height="5" style={{ color: "var(--colors-fg-default)" }} />
                {hasUnreadNotifications && (
                  <Box
                    position="absolute"
                    borderRadius="full"
                    w="2"
                    h="2"
                    style={{ top: "8px", right: "8px", backgroundColor: TRAYA_COLORS.heart }}
                  />
                )}
              </styled.div>
            </Link>

            <Link href="/search" aria-label="Search">
              <styled.div display="inline-flex" alignItems="center" justifyContent="center" w="10" h="10">
                <SearchIcon width="5" height="5" style={{ color: "var(--colors-fg-default)" }} />
              </styled.div>
            </Link>

            {session && <SelfAvatarBadge account={session} />}
          </HStack>
        </HStack>

        <VStack alignItems="start" gap="4" width="full" px="4" pt="3" pb="3">
          <styled.h1
            fontWeight="semibold"
            style={{ margin: "0", fontSize: "20px", lineHeight: "24px", letterSpacing: "-0.1px", color: "#2c2c2a" }}
          >
            {channel.name}
          </styled.h1>

          <ThreadCreateTrigger
            channelID={channel.id}
            channelName={channel.name}
            promptNudges={parsePromptNudges(channel.meta)}
          />

          {/* Collapses while reading. Kept mounted and clipped rather than
              unmounted so the composer keeps any half-typed draft. */}
          <Collapsible isCollapsed={isScrolledDown}>
            <UsersPostedToday signedIn={Boolean(session)} channelID={channel.id} />
          </Collapsible>
        </VStack>
      </VStack>

      {/* Filter Chips — hidden for admins on desktop (they use the admin Filters bar) */}
      <styled.div
        display={canManagePosts ? { base: "block", md: "none" } : "block"}
        width="full"
        bg="white"
        px="4"
      >
        <Collapsible isCollapsed={isScrolledDown} expandedPadding="12px 0">
          <FeedFilterChips
            selectedPrimaryTopic={selectedPrimaryTopic}
            onPrimaryTopicChange={onPrimaryTopicChange}
          />
        </Collapsible>
      </styled.div>
    </VStack>
  );
}

/**
 * Collapses its content to zero height when the reader scrolls down. Animates
 * grid-template-rows rather than height so the row resolves from the content and
 * no fixed pixel value has to be maintained per section.
 */
function Collapsible({
  isCollapsed,
  expandedPadding,
  children,
}: {
  isCollapsed: boolean;
  expandedPadding?: string;
  children: React.ReactNode;
}) {
  return (
    <styled.div
      width="full"
      style={{
        display: "grid",
        gridTemplateRows: isCollapsed ? "0fr" : "1fr",
        opacity: isCollapsed ? 0 : 1,
        padding: isCollapsed ? "0" : expandedPadding ?? "0",
        // One duration for every animated property. Staggering them makes the
        // document height settle after the fade, which the scroll listener then
        // reads as fresh movement.
        transition:
          "grid-template-rows 200ms ease, opacity 200ms ease, padding 200ms ease",
        willChange: "grid-template-rows",
      }}
    >
      <styled.div width="full" style={{ overflow: "hidden", minHeight: 0 }}>
        {children}
      </styled.div>
    </styled.div>
  );
}
