"use client";

import { useRouter } from "next/navigation";
import { ReactNode } from "react";

import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { styled } from "@/styled-system/jsx";

type Props = {
  title: ReactNode;
  onBack?: () => void;
};

export function CenteredBackHeader({ title, onBack }: Props) {
  const router = useRouter();
  const handleBack = onBack ?? (() => router.back());

  return (
    <styled.div
      position="sticky"
      top="0"
      zIndex="sticky"
      display="flex"
      alignItems="center"
      w="full"
      h="14"
      px="1"
      bg="bg.default"
    >
      <styled.button
        type="button"
        aria-label="Go back"
        onClick={handleBack}
        display="flex"
        alignItems="center"
        justifyContent="center"
        w="12"
        h="12"
        flexShrink="0"
        color="fg.default"
        bg="transparent"
        border="none"
        cursor="pointer"
      >
        <ArrowLeftIcon width="6" height="6" />
      </styled.button>

      <styled.h1
        flex="1"
        minW="0"
        fontSize="xl"
        lineHeight="[24px]"
        fontWeight="bold"
        letterSpacing="[0.2px]"
        color="fg.default"
        textAlign="center"
      >
        {title}
      </styled.h1>

      <styled.div w="12" flexShrink="0" />
    </styled.div>
  );
}
