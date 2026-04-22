"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthTrayaToken, usernameSet } from "@/api/openapi-client/auth";
import { useAccountGet } from "@/api/openapi-client/accounts";
import { channelList } from "@/api/openapi-client/channels";
import { styled } from "@/styled-system/jsx";
import { handle } from "@/api/client";
import { Button } from "@/components/ui/button";
// Icons commented out - Enter Community screen disabled
// import { VStack } from "@/styled-system/jsx";
// import { MembersIcon } from "@/components/ui/icons/Members";
// import { LikeIcon } from "@/components/ui/icons/Like";
// import { DiscussionIcon } from "@/components/ui/icons/Discussion";
// import { IntelligenceIcon } from "@/components/ui/icons/Intelligence";
// import { UsernameModal } from "../UsernameSelectionScreen/UsernameModal";
// import { useDisclosure } from "@/utils/useDisclosure";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { Spinner } from "@/components/ui/Spinner";
import { generateRandomUsername } from "@/utils/generateUsername";

type LandingScreenProps = {
  token: string;
  share?: boolean;
  streakCount?: number;
  rewardCoins?: number;
};

export function LandingScreen({ token, share, streakCount, rewardCoins }: LandingScreenProps) {
  const router = useRouter();
  const { trigger } = useAuthTrayaToken({ token });
  const { mutate: mutateAccount } = useAccountGet();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const triggered = useRef(false);
  // States commented out - Enter Community screen disabled (direct landing)
  // const [needsUsername, setNeedsUsername] = useState(false);
  // const [userName, setUserName] = useState<string | undefined>();
  // const usernameModal = useDisclosure();
  const { trackOnboardingLanded, trackEnterClicked } = useEventTracking();

  const buildShareQuery = () => {
    const params = new URLSearchParams();
    if (streakCount !== undefined) params.set("streak_count", String(streakCount));
    if (rewardCoins !== undefined) params.set("reward_coins", String(rewardCoins));
    const query = params.toString();
    return query ? `?${query}` : "";
  };

  const redirectToFirstChannel = async () => {
    const shareQuery = share ? buildShareQuery() : "";
    try {
      const channelsResponse = await channelList();
      const channels = channelsResponse?.channels ?? [];
      const monthChannel = channels.find((c) => c.name?.toLowerCase().includes("month"));
      const targetChannel = monthChannel ?? channels[0];
      if (targetChannel?.id) {
        const hasShareParams = share && streakCount !== undefined && rewardCoins !== undefined;
        if (hasShareParams) {
          router.push(`/channels/${targetChannel.id}/share-post${shareQuery}`);
        } else {
          router.push(`/channels/${targetChannel.id}`);
        }
      } else {
        router.push(`/channels`);
      }
    } catch {
      router.push(`/channels`);
    }
  };

  useEffect(() => {
    trackOnboardingLanded();
  }, [trackOnboardingLanded]);

  useEffect(() => {
    if (triggered.current) return;
    triggered.current = true;

    handle(
      async () => {
        await trigger();
        const account = await mutateAccount();
        const hasTempHandle = account?.handle?.startsWith("temp_");
        const hasNoHandle = !account?.handle;
        const isNewUser = hasNoHandle || hasTempHandle;

        if (isNewUser) {
          try {
            const randomUsername = generateRandomUsername(account?.name);
            await usernameSet({ username: randomUsername });
            await mutateAccount();
          } catch {
            // Silent fail - profile screen will handle it
          }
        }
        await redirectToFirstChannel();
      },
      {
        errorToast: false,
        onError: async () => {
          triggered.current = false;
          setIsLoading(false);
          setError("Authentication failed. Please try again.");
        },
      }
    );
  }, [trigger, router, mutateAccount, share]);

  // Retry handler for error state
  const handleEnterCommunity = async () => {
    trackEnterClicked();

    if (triggered.current) return;
    triggered.current = true;
    setIsLoading(true);
    setError(null);

    handle(
      async () => {
        await trigger();
        const account = await mutateAccount();
        const hasTempHandle = account?.handle?.startsWith("temp_");
        const hasNoHandle = !account?.handle;
        const isNewUser = hasNoHandle || hasTempHandle;

        if (isNewUser) {
          try {
            const randomUsername = generateRandomUsername(account?.name);
            await usernameSet({ username: randomUsername });
            await mutateAccount();
          } catch {
            // Silent fail - profile screen will handle it
          }
        }
        await redirectToFirstChannel();
      },
      {
        errorToast: false,
        onError: async () => {
          triggered.current = false;
          setIsLoading(false);
          setError("Authentication failed. Please try again.");
        },
      }
    );
  };

  // Direct landing - skip Enter Community screen, auto-redirect
  if (isLoading && !error) {
    return (
      <styled.div
        display="flex"
        flexDirection="column"
        alignItems="center"
        justifyContent="center"
        minH="screen"
        p="6"
        style={{ background: "#f5f5f5" }}
      >
        <Spinner size="lg" />
        <styled.p mt="4" color="fg.muted" fontSize="sm" textAlign="center">
          Hang tight, we're connecting you to your community!
        </styled.p>
      </styled.div>
    );
  }

  // Error state - show retry option
  if (error) {
    return (
      <styled.div
        display="flex"
        flexDirection="column"
        alignItems="center"
        justifyContent="center"
        minH="screen"
        p="6"
        style={{ background: "#f5f5f5" }}
      >
        <styled.div w="full" maxW="sm">
          <styled.div
            p="4"
            mb="6"
            rounded="lg"
            style={{ background: "#fee2e2", borderColor: "#fca5a5", borderWidth: "1px" }}
          >
            <styled.p fontSize="sm" style={{ color: "#dc2626" }}>
              {error}
            </styled.p>
          </styled.div>

          <Button
            w="full"
            onClick={handleEnterCommunity}
            loading={isLoading}
            disabled={isLoading}
            style={{
              background: "#4a9d6f",
              color: "#ffffff",
              padding: "12px 24px",
              borderRadius: "8px",
              fontSize: "16px",
              fontWeight: "500",
              border: "none",
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
          >
            {isLoading ? "Authenticating..." : "Try Again"}
          </Button>
        </styled.div>
      </styled.div>
    );
  }

  // Fallback loading state
  return (
    <styled.div
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      minH="screen"
      p="6"
      style={{ background: "#f5f5f5" }}
    >
      <Spinner size="lg" />
      <styled.p mt="4" color="fg.muted" fontSize="sm" textAlign="center">
        Hang tight, we're connecting you to your community!
      </styled.p>
    </styled.div>
  );

  /* Enter Community Screen - Commented out: direct landing instead
  return (
    <styled.div
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      minH="screen"
      p="6"
      style={{ background: "#f5f5f5" }}
    >
      <styled.div w="full" maxW="sm">
        <styled.div
          w="16"
          h="16"
          rounded="2xl"
          display="flex"
          alignItems="center"
          justifyContent="center"
          mx="auto"
          mb="8"
          style={{
            background: "#4a9d6f",
          }}
        >
          <MembersIcon width="8" height="8" style={{ color: "#ffffff" }} />
        </styled.div>

        <styled.h1
          fontSize="2xl"
          fontWeight="bold"
          color="fg.default"
          mb="3"
          textAlign="center"
        >
          Welcome to Your Community
        </styled.h1>

        <styled.p
          fontSize="sm"
          color="fg.muted"
          mb="8"
          textAlign="center"
        >
          Join thousands of hair warriors supporting each other on their
          transformation journey
        </styled.p>

        <VStack gap="3" mb="8" w="full">
          <styled.div
            display="flex"
            alignItems="center"
            gap="4"
            bg="white"
            rounded="lg"
            p="4"
            borderWidth="thin"
            borderColor="border.default"
            w="full"
            style={{ minHeight: "80px" }}
          >
            <styled.div
              w="10"
              h="10"
              rounded="lg"
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              style={{ background: "#f0f0f0" }}
            >
              <LikeIcon width="5" height="5" style={{ color: "#4a9d6f" }} />
            </styled.div>
            <VStack gap="0.5" w="full" alignItems="flex-start">
              <styled.h4 fontWeight="bold" color="fg.default" fontSize="sm" textAlign="left">
                Real Support
              </styled.h4>
              <styled.p fontSize="xs" color="fg.muted" style={{ lineHeight: "1.3" }} textAlign="left">
                Connect with others who truly understand
              </styled.p>
            </VStack>
          </styled.div>

          <styled.div
            display="flex"
            alignItems="center"
            gap="4"
            bg="white"
            rounded="lg"
            p="4"
            borderWidth="thin"
            borderColor="border.default"
            w="full"
            style={{ minHeight: "80px" }}
          >
            <styled.div
              w="10"
              h="10"
              rounded="lg"
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              style={{ background: "#f0f0f0" }}
            >
              <DiscussionIcon width="5" height="5" style={{ color: "#4a9d6f" }} />
            </styled.div>
            <VStack gap="0.5" w="full" alignItems="flex-start">
              <styled.h4 fontWeight="bold" color="fg.default" fontSize="sm" textAlign="left">
                Share Your Journey
              </styled.h4>
              <styled.p fontSize="xs" color="fg.muted" style={{ lineHeight: "1.3" }} textAlign="left">
                Exchange tips and celebrate progress together
              </styled.p>
            </VStack>
          </styled.div>

          <styled.div
            display="flex"
            alignItems="center"
            gap="4"
            bg="white"
            rounded="lg"
            p="4"
            borderWidth="thin"
            borderColor="border.default"
            w="full"
            style={{ minHeight: "80px" }}
          >
            <styled.div
              w="10"
              h="10"
              rounded="lg"
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              style={{ background: "#f0f0f0" }}
            >
              <IntelligenceIcon width="5" height="5" style={{ color: "#4a9d6f" }} />
            </styled.div>
            <VStack gap="0.5" w="full" alignItems="flex-start">
              <styled.h4 fontWeight="bold" color="fg.default" fontSize="sm" textAlign="left">
                Expert Access
              </styled.h4>
              <styled.p fontSize="xs" color="fg.muted" style={{ lineHeight: "1.3" }} textAlign="left">
                Get advice from Traya doctors and coaches
              </styled.p>
            </VStack>
          </styled.div>
        </VStack>

        {error && (
          <styled.div
            p="4"
            mb="6"
            rounded="lg"
            style={{ background: "#fee2e2", borderColor: "#fca5a5", borderWidth: "1px" }}
          >
            <styled.p fontSize="sm" style={{ color: "#dc2626" }}>
              {error}
            </styled.p>
          </styled.div>
        )}

        <Button
          w="full"
          onClick={handleEnterCommunity}
          loading={isLoading}
          disabled={isLoading}
          style={{
            background: "#4a9d6f",
            color: "#ffffff",
            padding: "12px 24px",
            borderRadius: "8px",
            fontSize: "16px",
            fontWeight: "500",
            border: "none",
            cursor: isLoading ? "not-allowed" : "pointer",
          }}
        >
          {isLoading ? "Authenticating..." : error ? "Try Again" : "Enter Community"}
        </Button>
      </styled.div>
    </styled.div>
  );
  */
}
