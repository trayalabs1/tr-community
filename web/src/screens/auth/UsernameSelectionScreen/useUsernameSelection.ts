import { useState, useEffect, useCallback } from "react";
import { usernameCheck, usernameSet } from "@/api/openapi-client/auth";
import { handle } from "@/api/client";

const USERNAME_PATTERN = /^[a-zA-Z0-9_-]+$/;
const MIN_LENGTH = 3;
const MAX_LENGTH = 30; // Matches backend validation limit

export function useUsernameSelection() {
  const [username, setUsername] = useState("");
  const [isChecking, setIsChecking] = useState(false);
  const [isAvailable, setIsAvailable] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [checkTimeout, setCheckTimeout] = useState<NodeJS.Timeout | null>(null);

  // Validate username format
  const validateUsername = (value: string): string | null => {
    if (value.length < MIN_LENGTH) {
      return `Username must be at least ${MIN_LENGTH} characters`;
    }
    if (value.length > MAX_LENGTH) {
      return `Username must be at most ${MAX_LENGTH} characters`;
    }
    if (!USERNAME_PATTERN.test(value)) {
      return "Username can only contain letters, numbers, underscores, and hyphens";
    }
    return null;
  };

  // Check username availability with debouncing (500ms)
  const checkAvailability = useCallback(async (value: string) => {
    const validationError = validateUsername(value);
    if (validationError) {
      setError(validationError);
      setIsAvailable(null);
      setIsChecking(false);
      return;
    }

    setError(null);
    setIsChecking(true);

    try {
      const result = await usernameCheck({ username: value });
      setIsAvailable(result.available);
      setIsChecking(false);
    } catch (err) {
      console.error("Username availability check failed:", err);
      setError("Failed to check username availability");
      setIsAvailable(null);
      setIsChecking(false);
    }
  }, []);

  // Debounce username checks
  useEffect(() => {
    if (!username) {
      setIsAvailable(null);
      setError(null);
      setIsChecking(false);
      return;
    }

    // Clear previous timeout
    if (checkTimeout) {
      clearTimeout(checkTimeout);
    }

    // Set new timeout for checking
    const timeout = setTimeout(() => {
      checkAvailability(username);
    }, 500); // 500ms debounce

    setCheckTimeout(timeout);

    return () => {
      if (timeout) {
        clearTimeout(timeout);
      }
    };
  }, [username, checkAvailability]);

  // Submit username
  const handleSubmit = async (): Promise<boolean> => {
    if (!username || isAvailable !== true) {
      return false;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      await handle(
        async () => {
          await usernameSet({ username });
        },
        {
          errorToast: false,
          onError: async (err) => {
            let errorMessage = "Failed to set username";

            if (err instanceof Error) {
              if ("status" in err) {
                const status = (err as any).status;
                if (status === 409) {
                  errorMessage = "Username is already taken";
                  setIsAvailable(false);
                } else if (status === 400) {
                  errorMessage = "Invalid username format";
                }
              } else {
                errorMessage = err.message || errorMessage;
              }
            }

            setError(errorMessage);
            setIsSubmitting(false);
          },
        },
      );
      return true;
    } catch (err) {
      setIsSubmitting(false);
      return false;
    }
  };

  return {
    username,
    setUsername,
    isChecking,
    isAvailable,
    error,
    isSubmitting,
    handleSubmit,
  };
}
