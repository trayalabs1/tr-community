"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSWRConfig } from "swr";

import { channelThreadCreate } from "src/api/openapi-client/channels";
import { ThreadInitialProps, Visibility, Permission } from "src/api/openapi-schema";

import { Button } from "@/components/ui/button";
import { useSession } from "@/auth";
import { hasPermission } from "@/utils/permissions";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { STREAK_IMAGES, DEFAULT_STREAK_IMAGE_KEY } from "@/lib/constants";
import { styled } from "@/styled-system/jsx";
import { VStack } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

type Props = {
  channelID: string;
  streakCount?: number;
  rewardCoins?: number;
};

function getStreakImageUrl(streakCount: number): string {
  return STREAK_IMAGES[streakCount] ?? STREAK_IMAGES[DEFAULT_STREAK_IMAGE_KEY] ?? "";
}

function buildStreakBody(streakCount: number, rewardCoins: number): string {
  const imageUrl = getStreakImageUrl(streakCount);
  return `<p>${streakCount} days streak completed 🔥</p><img src="${imageUrl}" alt="${streakCount}-Day streak - Won ${rewardCoins} coins" />`;
}

export function SharePostScreen({ channelID, streakCount, rewardCoins }: Props) {
  const router = useRouter();
  const { mutate } = useSWRConfig();
  const session = useSession();
  const { trackSubmitForReview } = useEventTracking();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const hasValidParams = streakCount !== undefined && rewardCoins !== undefined;

  if (!hasValidParams) {
    router.replace(`/channels/${channelID}`);
    return null;
  }

  const imageUrl = getStreakImageUrl(streakCount);

  const handlePost = async () => {
    if (isSubmitting) return;

    setIsSubmitting(true);
    setError(null);

    try {
      const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);
      const targetVisibility = isAdmin ? Visibility.published : Visibility.review;
      const body = buildStreakBody(streakCount, rewardCoins);

      trackSubmitForReview(body.length, false, false, channelID);

      const payload: ThreadInitialProps = {
        title: "",
        body,
        visibility: targetVisibility,
      };

      await channelThreadCreate(channelID, payload);

      mutate(
        (key: unknown) =>
          Array.isArray(key) &&
          typeof key[0] === "string" &&
          key[0] === `/channels/${channelID}/threads`,
      );

      router.replace(`/channels/${channelID}`);
    } catch {
      setError("Failed to create post. Please try again.");
      setIsSubmitting(false);
    }
  };

  return (
    <VStack gap="0" minH="dvh" style={{ backgroundColor: "#f9fafb" }}>
      <styled.div
        w="full"
        py="3"
        px="4"
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        style={{
          backgroundColor: "white",
          borderBottom: "1px solid #e5e7eb",
        }}
      >
        <styled.button
          type="button"
          onClick={() => router.replace(`/channels/${channelID}`)}
          fontSize="sm"
          cursor="pointer"
          style={{
            background: "none",
            border: "none",
            color: TRAYA_COLORS.primary,
            fontWeight: 500,
          }}
        >
          Skip
        </styled.button>

        <styled.span fontSize="md" fontWeight="semibold">
          Share Your Streak
        </styled.span>

        <styled.div style={{ width: "40px" }} />
      </styled.div>

      <VStack gap="6" flex="1" w="full" p="4" maxW="lg" mx="auto" alignItems="center" justifyContent="center">
        <VStack
          gap="4"
          w="full"
          p="4"
          style={{
            backgroundColor: "white",
            borderRadius: "16px",
            boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
          }}
        >
          <styled.p fontSize="lg" fontWeight="semibold" textAlign="center">
            {streakCount} days streak completed 🔥
          </styled.p>

          <styled.img
            src={imageUrl}
            alt={`${streakCount}-Day streak - Won ${rewardCoins} coins`}
            w="full"
            style={{ borderRadius: "12px" }}
          />
        </VStack>

        {error && (
          <styled.p fontSize="sm" style={{ color: "#dc2626" }}>
            {error}
          </styled.p>
        )}

        <VStack gap="3" w="full">
          <Button
            w="full"
            size="lg"
            onClick={handlePost}
            loading={isSubmitting}
            disabled={isSubmitting}
            style={{
              backgroundColor: TRAYA_COLORS.primary,
              color: "white",
            }}
          >
            Share with Community
          </Button>

          <styled.button
            type="button"
            onClick={() => router.replace(`/channels/${channelID}`)}
            w="full"
            py="3"
            fontSize="sm"
            cursor="pointer"
            style={{
              background: "none",
              border: "none",
              color: "#6b7280",
            }}
          >
            Maybe later
          </styled.button>
        </VStack>
      </VStack>
    </VStack>
  );
}
