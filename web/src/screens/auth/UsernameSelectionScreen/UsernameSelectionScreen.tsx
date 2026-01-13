"use client";

import { useRouter } from "next/navigation";
import { VStack, styled } from "@/styled-system/jsx";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MembersIcon } from "@/components/ui/icons/Members";
import { useUsernameSelection } from "./useUsernameSelection";

export function UsernameSelectionScreen() {
  const router = useRouter();
  const {
    username,
    setUsername,
    isChecking,
    isAvailable,
    error,
    isSubmitting,
    handleSubmit,
  } = useUsernameSelection();

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const success = await handleSubmit();
    if (success) {
      router.push("/");
    }
  };

  // Validation state for UI feedback
  const showValidation = username.length > 0;
  const validationColor = error
    ? "#dc2626"
    : isAvailable === true
      ? "#16a34a"
      : isAvailable === false
        ? "#dc2626"
        : "#6b7280";

  const validationMessage = error
    ? error
    : isChecking
      ? "Checking availability..."
      : isAvailable === true
        ? "✓ Username is available"
        : isAvailable === false
          ? "✗ Username is already taken"
          : username.length >= 3
            ? "• 3-30 characters, letters, numbers, _ and - only"
            : "• At least 3 characters required";

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
          Choose Your Username
        </styled.h1>

        {/* Description */}
        <styled.p
          fontSize="sm"
          color="fg.muted"
          mb="8"
          textAlign="center"
        >
          This is how you'll be known in the community. Choose wisely!
        </styled.p>

        {/* Username Form */}
        <form onSubmit={handleFormSubmit}>
          <VStack gap="4" w="full">
            <styled.div w="full">
              <Input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter your username"
                disabled={isSubmitting}
                autoFocus
                autoComplete="off"
                spellCheck={false}
                style={{
                  fontSize: "16px",
                  padding: "12px 16px",
                  width: "100%",
                  borderColor: showValidation ? validationColor : undefined,
                }}
              />

              {/* Validation Message */}
              {showValidation && (
                <styled.p
                  fontSize="xs"
                  mt="2"
                  style={{
                    color: validationColor,
                  }}
                >
                  {validationMessage}
                </styled.p>
              )}
            </styled.div>

            {/* Submit Button */}
            <Button
              type="submit"
              w="full"
              loading={isSubmitting}
              disabled={
                isSubmitting || !username || isAvailable !== true || !!error
              }
              style={{
                background: "#4a9d6f",
                color: "#ffffff",
                padding: "12px 24px",
                borderRadius: "8px",
                fontSize: "16px",
                fontWeight: "500",
                border: "none",
                cursor:
                  isSubmitting || !username || isAvailable !== true || !!error
                    ? "not-allowed"
                    : "pointer",
                opacity:
                  isSubmitting || !username || isAvailable !== true || !!error
                    ? 0.5
                    : 1,
              }}
            >
              {isSubmitting ? "Setting username..." : "Continue"}
            </Button>
          </VStack>
        </form>

        {/* Help Text */}
        <styled.p
          fontSize="xs"
          color="fg.muted"
          mt="6"
          textAlign="center"
        >
          Your username must be unique and can't be changed later
        </styled.p>
      </styled.div>
    </styled.div>
  );
}
