"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import { useAuthTrayaToken } from "@/api/openapi-client/auth";
import { VStack, styled } from "@/styled-system/jsx";
import { handle } from "@/api/client";
import { Button } from "@/components/ui/button";
import { MembersIcon } from "@/components/ui/icons/Members";
import { LikeIcon } from "@/components/ui/icons/Like";
import { DiscussionIcon } from "@/components/ui/icons/Discussion";
import { IntelligenceIcon } from "@/components/ui/icons/Intelligence";

export function LandingScreen({ token }: { token: string }) {
  const router = useRouter();
  const { trigger } = useAuthTrayaToken({ token });
  const [isLoading, setIsLoading] = useState(false);
  const triggered = useRef(false);

  const handleJoinCommunity = async () => {
    if (triggered.current) return;
    triggered.current = true;
    setIsLoading(true);

    handle(
      async () => {
        await trigger();
        router.push("/");
      },
      {
        onError: async () => {
          triggered.current = false;
          setIsLoading(false);
          router.push("/login");
        },
      }
    );
  };

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
        {/* Icon Container */}
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

        {/* Heading */}
        <styled.h1
          fontSize="2xl"
          fontWeight="bold"
          color="fg.default"
          mb="3"
          textAlign="center"
        >
          Welcome to Traya Community
        </styled.h1>

        {/* Description */}
        <styled.p
          fontSize="sm"
          color="fg.muted"
          mb="8"
          textAlign="center"
        >
          Join thousands of hair warriors supporting each other on their
          transformation journey
        </styled.p>

        {/* Feature Cards */}
        <VStack gap="3" mb="8" w="full">
          {/* Real Support Card */}
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

          {/* Share Your Journey Card */}
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

          {/* Expert Access Card */}
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

        {/* Join Button */}
        <Button
          w="full"
          onClick={handleJoinCommunity}
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
          {isLoading ? "Authenticating..." : "Join Community"}
        </Button>
      </styled.div>
    </styled.div>
  );
}
