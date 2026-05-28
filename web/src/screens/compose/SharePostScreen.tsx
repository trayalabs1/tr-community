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
import {
  STREAK_IMAGES,
  DEFAULT_STREAK_IMAGE_KEY,
  FEEDBACK_PROGRESS_IMAGE,
} from "@/lib/constants";
import { styled } from "@/styled-system/jsx";
import { VStack } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

type Props = {
  channelID: string;
  streakCount?: number;
  rewardCoins?: number;
  category?: string;
  type?: string;
};

function getStreakImageUrl(streakCount: number): string {
  return STREAK_IMAGES[streakCount] ?? STREAK_IMAGES[DEFAULT_STREAK_IMAGE_KEY] ?? "";
}

function buildStreakBody(streakCount: number, rewardCoins: number): string {
  const imageUrl = getStreakImageUrl(streakCount);
  return `<p>${streakCount} days streak completed 🔥</p><img src="${imageUrl}" alt="${streakCount}-Day streak - Won ${rewardCoins} coins" />`;
}

const FEEDBACK_PROGRESS_COPY =
  "I'm seeing fewer strands every day, and I feel so much better about my hair! 🙌";

function buildFeedbackBody(): string {
  return `<p>${FEEDBACK_PROGRESS_COPY}</p><img src="${FEEDBACK_PROGRESS_IMAGE}" alt="Progress update" />`;
}

export function SharePostScreen({ channelID, streakCount, rewardCoins, category, type }: Props) {
  const router = useRouter();
  const { mutate } = useSWRConfig();
  const session = useSession();
  const { trackSubmitForReview, trackSharePostCommunity } = useEventTracking();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isFeedbackProgress = category === "feedback" && type === "progress";
  const isBahStreak = streakCount !== undefined && rewardCoins !== undefined;

  if (!isFeedbackProgress && !isBahStreak) {
    router.replace(`/channels/${channelID}`);
    return null;
  }

  const previewImage = isFeedbackProgress
    ? FEEDBACK_PROGRESS_IMAGE
    : getStreakImageUrl(streakCount as number);
  const previewCopy = isFeedbackProgress
    ? FEEDBACK_PROGRESS_COPY
    : `I just won ${rewardCoins} coins by completing my ${streakCount} day streak  🙌 `;
  const previewAlt = isFeedbackProgress
    ? "Progress update"
    : `${streakCount}-Day streak - Won ${rewardCoins} coins`;

  const handlePost = async () => {
    if (isSubmitting) return;

    setIsSubmitting(true);
    setError(null);

    try {
      const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);
      const targetVisibility = isFeedbackProgress
        ? Visibility.review
        : isAdmin
          ? Visibility.published
          : Visibility.review;
      const body = isFeedbackProgress
        ? buildFeedbackBody()
        : buildStreakBody(streakCount as number, rewardCoins as number);

      if (isFeedbackProgress) {
        trackSharePostCommunity("progress_update", channelID);
      } else {
        trackSubmitForReview(body.length, false, false, channelID);
      }

      const meta = isFeedbackProgress
        ? { post_category: "feedback", type: "progress" }
        : { post_category: "BAH", type: streakCount };

      const payload: ThreadInitialProps = {
        title: "",
        body,
        visibility: targetVisibility,
        meta,
      };

      await channelThreadCreate(channelID, payload);

      await mutate(
        (key: unknown) =>
          Array.isArray(key) &&
          typeof key[0] === "string" &&
          key[0].startsWith(`/channels/${channelID}/threads`),
      );

      router.replace(`/channels/${channelID}`);
    } catch {
      setError("Failed to create post. Please try again.");
      setIsSubmitting(false);
    }
  };

  const headerTitle = isFeedbackProgress ? "Share Your Progress" : "Share Your Streak";

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
          {headerTitle}
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
            {previewCopy}
          </styled.p>

          <styled.img src={previewImage} alt={previewAlt} w="full" />
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
