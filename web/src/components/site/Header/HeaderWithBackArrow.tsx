"use client";

import { useRouter } from "next/navigation";
import { ReactNode } from "react";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { HStack, VStack, styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

interface HeaderWithBackArrowProps {
  title: ReactNode;
  subtitle?: ReactNode;
  headerIcon?: ReactNode;
  headerIconBackground?: string;
  onBack?: () => void;
  mobileOnly?: boolean;
  isSticky?: boolean;
  showBorder?: boolean;
  containerProps?: React.ComponentProps<typeof HStack>;
}

const BACK_BUTTON_STYLES = {
  marginLeft: "-0.5rem",
  background: "transparent",
  border: "none",
  cursor: "pointer",
  borderRadius: "0.75rem",
  transition: "background-color 0.2s",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
} as const;

const STICKY_STYLES = {
  position: "sticky" as const,
  top: 0,
  zIndex: 10,
  backgroundColor: "rgba(255, 255, 255, 0.95)",
  backdropFilter: "blur(12px)",
  flexShrink: 0,
  borderBottomColor: "rgba(0, 0, 0, 0.05)",
  margin: 0,
};

export function HeaderWithBackArrow({
  title,
  subtitle,
  headerIcon,
  headerIconBackground,
  onBack,
  mobileOnly = false,
  isSticky = mobileOnly,
  showBorder = true,
  containerProps,
}: HeaderWithBackArrowProps) {
  const router = useRouter();
  const handleBack = onBack || (() => router.back());

  const containerStyle: React.CSSProperties = isSticky ? STICKY_STYLES : {};

  const headerContent = (
    <>
      <styled.button
        onClick={handleBack}
        p="2"
        style={BACK_BUTTON_STYLES}
        onMouseEnter={(e) => {
          (e.currentTarget as HTMLElement).style.backgroundColor = "rgba(0, 0, 0, 0.05)";
        }}
        onMouseLeave={(e) => {
          (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
        }}
        aria-label="Go back"
      >
        <ArrowLeftIcon width="5" height="5" />
      </styled.button>

      {headerIcon && (
        <styled.div
          w="8"
          h="8"
          minW="8"
          minH="8"
          rounded="full"
          display="flex"
          alignItems="center"
          justifyContent="center"
          flexShrink="0"
          style={{
            backgroundColor: headerIconBackground || `${TRAYA_COLORS.primary}`,
          }}
        >
          {headerIcon}
        </styled.div>
      )}

      <VStack gap="0" alignItems="start" width="full" minWidth="0">
        <styled.h1 fontSize="md" fontWeight="semibold" color="fg.default">
          {title}
        </styled.h1>
        {subtitle && (
          <styled.p fontSize="xs" color="fg.muted">
            {subtitle}
          </styled.p>
        )}
      </VStack>
    </>
  );

  const displayValue = mobileOnly
    ? ({ base: "flex", md: "none" } as any)
    : ("flex" as any);
  const borderBottomWidth = (showBorder ? "thin" : "none") as any;

  const hstackProps: any = {
    gap: "2",
    alignItems: "center",
    width: "full",
    m: "0",
    mt: "0",
    ml: "0",
    mr: "0",
    mb: "0",
    p: "4",
    borderBottomWidth,
    borderBottomColor: showBorder ? "border.default" : undefined,
    display: displayValue,
    style: containerStyle,
    ...containerProps,
  };

  return <HStack {...hstackProps}>{headerContent}</HStack>;
}
