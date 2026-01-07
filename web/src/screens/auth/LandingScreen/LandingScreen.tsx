"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import { useAuthTrayaToken } from "@/api/openapi-client/auth";
import { Center, VStack, styled } from "@/styled-system/jsx";
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
    <Center h="screen" p="6" style={{ background: "#f5f5f5" }}>
      <styled.div maxW="md" textAlign="center">
        {/* Icon Container */}
        <styled.div
          w="20"
          h="20"
          rounded="2xl"
          display="flex"
          alignItems="center"
          justifyContent="center"
          mx="auto"
          mb="6"
          style={{
            background: "#4a9d6f",
          }}
        >
          <MembersIcon width="10" height="10" style={{ color: "#ffffff" }} />
        </styled.div>

        {/* Heading */}
        <styled.h1 fontSize="2xl" fontWeight="bold" color="fg.default" mb="3">
          Welcome to Traya Community
        </styled.h1>

        {/* Description */}
        <styled.p color="fg.muted" mb="8">
          Join thousands of hair warriors supporting each other on their
          transformation journey
        </styled.p>

        {/* Feature Cards */}
        <VStack gap="4" mb="8">
          {/* Real Support Card */}
          <styled.div
            display="flex"
            alignItems="flex-start"
            gap="3"
            bg="bg.subtle"
            rounded="xl"
            p="4"
            borderWidth="thin"
            borderColor="border.default"
            textAlign="left"
          >
            <styled.div
              w="10"
              h="10"
              rounded="xl"
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              style={{ background: "#f0f0f0" }}
            >
              <LikeIcon width="5" height="5" style={{ color: "#4a9d6f" }} />
            </styled.div>
            <VStack gap="1">
              <styled.h4 fontWeight="medium" color="fg.default">
                Real Support
              </styled.h4>
              <styled.p fontSize="sm" color="fg.muted">
                Connect with others who truly understand
              </styled.p>
            </VStack>
          </styled.div>

          {/* Share Your Journey Card */}
          <styled.div
            display="flex"
            alignItems="flex-start"
            gap="3"
            bg="bg.subtle"
            rounded="xl"
            p="4"
            borderWidth="thin"
            borderColor="border.default"
            textAlign="left"
          >
            <styled.div
              w="10"
              h="10"
              rounded="xl"
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              style={{ background: "#f0f0f0" }}
            >
              <DiscussionIcon width="5" height="5" style={{ color: "#4a9d6f" }} />
            </styled.div>
            <VStack gap="1">
              <styled.h4 fontWeight="medium" color="fg.default">
                Share Your Journey
              </styled.h4>
              <styled.p fontSize="sm" color="fg.muted">
                Exchange tips and celebrate progress together
              </styled.p>
            </VStack>
          </styled.div>

          {/* Expert Access Card */}
          <styled.div
            display="flex"
            alignItems="flex-start"
            gap="3"
            bg="bg.subtle"
            rounded="xl"
            p="4"
            borderWidth="thin"
            borderColor="border.default"
            textAlign="left"
          >
            <styled.div
              w="10"
              h="10"
              rounded="xl"
              display="flex"
              alignItems="center"
              justifyContent="center"
              flexShrink="0"
              style={{ background: "#f0f0f0" }}
            >
              <IntelligenceIcon width="5" height="5" style={{ color: "#4a9d6f" }} />
            </styled.div>
            <VStack gap="1">
              <styled.h4 fontWeight="medium" color="fg.default">
                Expert Access
              </styled.h4>
              <styled.p fontSize="sm" color="fg.muted">
                Get advice from Traya doctors and coaches
              </styled.p>
            </VStack>
          </styled.div>
        </VStack>

        {/* Join Button */}
        <Button
          w="full"
          size="lg"
          onClick={handleJoinCommunity}
          loading={isLoading}
          disabled={isLoading}
          style={{
            background: "#4a9d6f",
            color: "#ffffff",
          }}
        >
          {isLoading ? "Authenticating..." : "Join Community"}
        </Button>
      </styled.div>
    </Center>
  );
}
