"use client";

import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { useAccountGet } from "@/api/openapi-client/accounts";
import { usernameSet } from "@/api/openapi-client/auth";
import { generateRandomUsername } from "@/utils/generateUsername";
import { styled } from "@/styled-system/jsx";
import { Spinner } from "@/components/ui/Spinner";

export function TempHandlePrompt() {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const { data: account, isLoading, error } = useAccountGet({
    swr: {
      revalidateOnMount: true,
    },
  });
  const triggered = useRef(false);

  useEffect(() => {
    if (triggered.current) return;

    if (error) {
      setErrorMessage("Something went wrong. Please try again later.");
      return;
    }

    if (isLoading) return;

    if (!account) {
      setErrorMessage("Something went wrong. Please try again later.");
      return;
    }

    const hasValidHandle = account.handle && !account.handle.startsWith("temp_");

    if (hasValidHandle) {
      router.replace(`/m/${account.handle}`);
      return;
    }

    triggered.current = true;

    const setUsernameAndRedirect = async () => {
      try {
        const randomUsername = generateRandomUsername(account.name);
        await usernameSet({ username: randomUsername });
        router.push(`/m/${randomUsername}`);
      } catch {
        setErrorMessage("Something went wrong. Please try again later.");
      }
    };

    setUsernameAndRedirect();
  }, [account, isLoading, error, router]);

  if (errorMessage) {
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
        <styled.div
          p="4"
          rounded="lg"
          style={{ background: "#fee2e2", borderColor: "#fca5a5", borderWidth: "1px" }}
        >
          <styled.p fontSize="sm" style={{ color: "#dc2626" }}>
            {errorMessage}
          </styled.p>
        </styled.div>
      </styled.div>
    );
  }

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
      <styled.p mt="4" color="fg.muted" fontSize="sm">
        Setting up your profile...
      </styled.p>
    </styled.div>
  );
}
