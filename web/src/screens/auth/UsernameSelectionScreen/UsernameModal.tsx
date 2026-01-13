"use client";

import { ModalDrawer } from "@/components/site/Modaldrawer/Modaldrawer";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MembersIcon } from "@/components/ui/icons/Members";
import { UseDisclosureProps } from "@/utils/useDisclosure";
import { VStack, styled } from "@/styled-system/jsx";
import { useUsernameSelection } from "./useUsernameSelection";

type Props = UseDisclosureProps & {
  onSuccess?: () => void;
};

export function UsernameModal({ onSuccess, ...disclosureProps }: Props) {
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
      onSuccess?.();
      disclosureProps.onClose?.();
    }
  };

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
    <>
      <ModalDrawer dismissable={!isSubmitting} {...disclosureProps}>
        <form onSubmit={handleFormSubmit}>
        <VStack gap="6" w="full" py="2">
          {/* Icon Container */}
          <styled.div
            w="16"
            h="16"
            rounded="2xl"
            display="flex"
            alignItems="center"
            justifyContent="center"
            mx="auto"
            style={{
              background: "#4a9d6f",
            }}
          >
            <MembersIcon width="8" height="8" style={{ color: "#ffffff" }} />
          </styled.div>

          {/* Heading */}
          <styled.h2
            fontSize="2xl"
            fontWeight="bold"
            color="fg.default"
            textAlign="center"
            m="0"
          >
            What should we call you?
          </styled.h2>

          {/* Description */}
          <styled.p
            fontSize="sm"
            color="fg.muted"
            textAlign="center"
            m="0"
          >
            This name will be visible on all your posts, comments, and profile in the community
          </styled.p>

          {/* Form Field */}
          <VStack gap="2" w="full">
            <styled.label
              fontSize="sm"
              fontWeight="medium"
              color="fg.default"
              display="block"
            >
              Display Name
            </styled.label>

            <Input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your name..."
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

            {/* Helper Text */}
            <styled.p
              fontSize="xs"
              color="fg.muted"
              m="0"
            >
              You can use your real name or a nickname
            </styled.p>

            {/* Validation Message */}
            {showValidation && (
              <styled.p
                fontSize="xs"
                m="0"
                style={{
                  color: validationColor,
                }}
              >
                {validationMessage}
              </styled.p>
            )}
          </VStack>

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
      </ModalDrawer>

      <style jsx global>{`
        .modaldrawer__overlay {
          background-color: rgba(0, 0, 0, 0.7) !important;
        }
      `}</style>
    </>
  );
}
