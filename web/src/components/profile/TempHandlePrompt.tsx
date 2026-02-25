"use client";

import { useRouter } from "next/navigation";
import { useCallback, useState, useEffect } from "react";
import { useAccountGet } from "@/api/openapi-client/accounts";
import { UsernameModal } from "@/screens/auth/UsernameSelectionScreen/UsernameModal";
import { styled } from "@/styled-system/jsx";
import { Spinner } from "@/components/ui/Spinner";

export function TempHandlePrompt() {
  const router = useRouter();
  const { data: account, isLoading, mutate: mutateAccount } = useAccountGet({
    swr: {
      revalidateOnMount: true,
    },
  });
  const [showModal, setShowModal] = useState(false);
  const [isRedirecting, setIsRedirecting] = useState(false);

  const hasValidHandle = account?.handle && !account.handle.startsWith("temp_");

  useEffect(() => {
    if (isLoading || isRedirecting) return;

    if (hasValidHandle) {
      setIsRedirecting(true);
      router.replace(`/m/${account.handle}`);
    } else if (account && !showModal) {
      setShowModal(true);
    }
  }, [account, hasValidHandle, isLoading, isRedirecting, router, showModal]);

  const handleSuccess = useCallback(async () => {
    setIsRedirecting(true);
    const updatedAccount = await mutateAccount();
    if (updatedAccount?.handle) {
      router.replace(`/m/${updatedAccount.handle}`);
    } else {
      router.push("/channels");
    }
  }, [mutateAccount, router]);

  const handleClose = useCallback(() => {
    setShowModal(false);
    router.push("/channels");
  }, [router]);

  const handleOpenChange = useCallback((e: { open: boolean }) => {
    if (!e.open) {
      handleClose();
    }
  }, [handleClose]);

  if (isLoading || isRedirecting || hasValidHandle) {
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
          Loading profile...
        </styled.p>
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
      <UsernameModal
        isOpen={showModal}
        onOpen={() => setShowModal(true)}
        onClose={handleClose}
        onOpenChange={handleOpenChange}
        onSuccess={handleSuccess}
        initialName={account?.name}
      />
    </styled.div>
  );
}
