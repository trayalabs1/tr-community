"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { useAuthTrayaToken } from "@/api/openapi-client/auth";
import { Center, VStack, styled } from "@/styled-system/jsx";
import { handle } from "@/api/client";
import { Spinner } from "@/components/ui/Spinner";

export function LandingScreen({ token }: { token: string }) {
  const router = useRouter();
  const { trigger } = useAuthTrayaToken({ token });
  const triggered = useRef(false);

  useEffect(() => {
    if (triggered.current) return;
    triggered.current = true;

    handle(async () => {
      await trigger();
      router.push("/");
    }, {
      onError: async () => {
        router.push("/login");
      }
    });
  }, [trigger, router]);

  return (
    <Center h="screen">
      <VStack gap="4">
        <Spinner size="lg" />
        <styled.p fontWeight="medium">Authenticating...</styled.p>
      </VStack>
    </Center>
  );
}
