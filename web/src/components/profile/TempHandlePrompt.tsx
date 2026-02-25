"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect } from "react";
import { useDisclosure } from "@/utils/useDisclosure";
import { UsernameModal } from "@/screens/auth/UsernameSelectionScreen/UsernameModal";
import { useAccountGet } from "@/api/openapi-client/accounts";

type Props = {
  initialName?: string;
};

export function TempHandlePrompt({ initialName }: Props) {
  const router = useRouter();
  const { mutate: mutateAccount } = useAccountGet();
  const usernameModal = useDisclosure();

  useEffect(() => {
    usernameModal.onOpen();
  }, []);

  const handleSuccess = useCallback(async () => {
    const account = await mutateAccount();
    if (account?.handle) {
      router.replace(`/m/${account.handle}`);
    }
  }, [mutateAccount, router]);

  return (
    <UsernameModal
      isOpen={usernameModal.isOpen}
      onOpen={usernameModal.onOpen}
      onClose={usernameModal.onClose}
      onOpenChange={usernameModal.onOpenChange}
      onSuccess={handleSuccess}
      initialName={initialName}
    />
  );
}
