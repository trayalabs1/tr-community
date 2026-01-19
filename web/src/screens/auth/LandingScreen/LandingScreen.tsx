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
import { UsernameModal } from "../UsernameSelectionScreen/UsernameModal";
import { useDisclosure } from "@/utils/useDisclosure";

export function LandingScreen({ token }: { token: string }) {
  const router = useRouter();
  const { trigger } = useAuthTrayaToken({ token });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const triggered = useRef(false);
  const usernameModal = useDisclosure();

  const handleJoinCommunity = async () => {
    if (triggered.current) return;
    triggered.current = true;
    setIsLoading(true);
    setError(null);

    handle(
      async () => {
        const response = await trigger();

        // Check if user needs to set a username
        if (response?.needs_username) {
          usernameModal.onOpen();
          // Keep loading state true while modal is open
        } else {
          // User doesn't need to set username, redirect to home
          setIsLoading(false);
          router.push("/");
        }
      },
      {
        errorToast: false,
        onError: async (err) => {

          let errorMessage = "Unknown error";
          let statusCode = undefined;

          if (err instanceof Error) {
            errorMessage = err.message;
            if ("status" in err) {
              statusCode = (err as any).status;
            }
          } else if (typeof err === "string") {
            errorMessage = err;
          } else {
            errorMessage = JSON.stringify(err);
          }

          triggered.current = false;
          setIsLoading(false);
          setError("Authentication failed. Please try again.");
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

        {/* Error Message */}
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

      {/* Username Modal */}
      <UsernameModal
        isOpen={usernameModal.isOpen}
        onOpen={usernameModal.onOpen}
        onClose={() => {
          // Reset the triggered flag and loading state if modal is dismissed
          triggered.current = false;
          setIsLoading(false);
          usernameModal.onClose();
        }}
        onOpenChange={usernameModal.onOpenChange}
        onSuccess={() => {
          setIsLoading(false);
          router.push("/");
        }}
      />
    </styled.div>
  );
}
