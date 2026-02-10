"use client";

import { useCallback } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { HStack, VStack, styled } from "@/styled-system/jsx";
import { SearchIcon } from "@/components/ui/icons/Search";
import { HomeIcon } from "@heroicons/react/24/outline";
import { LibraryIcon } from "@/components/ui/icons/Library";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import { TRAYA_COLORS } from "@/theme/traya-colors";
import { useEventTracking } from "@/lib/moengage/useEventTracking";

export function MobileCommandBar() {
  const pathname = usePathname();
  const { trackSearchClicked, trackInfoClicked } = useEventTracking();

  const isHomeActive = pathname === "/" || (pathname.startsWith("/channels") && !pathname.includes("/settings"));
  const isSearchActive = pathname.startsWith("/search");
  const isLibraryActive = pathname.startsWith("/l");
  const isInfoActive = pathname.startsWith("/info");

  const handleSearchClick = useCallback(() => {
    trackSearchClicked();
  }, [trackSearchClicked]);

  const handleInfoClick = useCallback(() => {
    trackInfoClicked();
  }, [trackInfoClicked]);

  const TabItem = ({
    href,
    icon: Icon,
    label,
    isActive,
    onClick,
  }: {
    href: string;
    icon: React.ComponentType<any>;
    label: string;
    isActive: boolean;
    onClick?: () => void;
  }) => (
    <Link href={href} style={{ textDecoration: "none" }} onClick={onClick}>
      <VStack
        alignItems="center"
        justifyContent="center"
        gap="1"
        flex="1"
        cursor="pointer"
        as="div"
      >
        <styled.div
          display="flex"
          alignItems="center"
          justifyContent="center"
          width="6"
          height="6"
          style={{
            color: isActive ? TRAYA_COLORS.primary : TRAYA_COLORS.neutral.textMuted,
          }}
        >
          <Icon width="24" height="24" strokeWidth={1.5} />
        </styled.div>
        <styled.span
          fontSize="xs"
          fontWeight="medium"
          style={{
            margin: "0",
            textAlign: "center",
            color: isActive ? TRAYA_COLORS.primary : TRAYA_COLORS.neutral.textMuted,
          }}
        >
          {label}
        </styled.span>
      </VStack>
    </Link>
  );

  return (
    <styled.nav
      position="fixed"
      bottom="0"
      left="0"
      right="0"
      width="full"
      backgroundColor="white"
      borderTopWidth="thin"
      borderTopColor="border.default"
      display={{ base: "flex", md: "none" }}
      style={{
        justifyContent: "center",
        alignItems: "flex-end",
        zIndex: 40,
        pointerEvents: "auto",
      }}
    >
      <HStack
        width="full"
        justifyContent="space-around"
        alignItems="stretch"
        py="3"
        px="0"
      >
        <TabItem
          href="/"
          icon={HomeIcon}
          label="Home"
          isActive={isHomeActive}
        />
        <TabItem
          href="/search"
          icon={SearchIcon}
          label="Search"
          isActive={isSearchActive}
          onClick={handleSearchClick}
        />
        {/* <TabItem
          href="/l"
          icon={LibraryIcon}
          label="Library"
          isActive={isLibraryActive}
        /> */}
        <TabItem
          href="/info"
          icon={InformationCircleIcon}
          label="Info"
          isActive={isInfoActive}
          onClick={handleInfoClick}
        />
      </HStack>
    </styled.nav>
  );
}
