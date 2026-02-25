"use client";

import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import { UsernameModal } from "@/screens/auth/UsernameSelectionScreen/UsernameModal";
import { useAccountGet } from "@/api/openapi-client/accounts";

type Props = {
  initialName?: string;
};

export function TempHandlePrompt({ initialName }: Props) {
  const router = useRouter();
  const { mutate: mutateAccount } = useAccountGet();
  const [isOpen, setIsOpen] = useState(true);

  const handleSuccess = useCallback(async () => {
    const account = await mutateAccount();
    if (account?.handle) {
      router.replace(`/m/${account.handle}`);
    }
  }, [mutateAccount, router]);

  const handleOpenChange = useCallback((e: { open: boolean }) => {
    setIsOpen(e.open);
  }, []);

  return (
    <UsernameModal
      isOpen={isOpen}
      onOpen={() => setIsOpen(true)}
      onClose={() => setIsOpen(false)}
      onOpenChange={handleOpenChange}
      onSuccess={handleSuccess}
      initialName={initialName}
    />
  );
}
