"use client";

import { Dialog, Portal } from "@ark-ui/react";
import { useState } from "react";

import { VStack, styled } from "@/styled-system/jsx";

type Props = {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void | Promise<void>;
  /** Question shown to the author, e.g. "…delete this comment?". */
  title?: string;
  /** Supporting line under the question. */
  message?: string;
  /** Label for the destructive action. */
  confirmLabel?: string;
  /** Label shown while the destructive action runs. */
  confirmingLabel?: string;
  /** Label for the action that keeps the content. */
  cancelLabel?: string;
};

export function DeleteConfirmSheet({
  isOpen,
  onOpenChange,
  onConfirm,
  title = "Are you sure you want to delete this post?",
  message = "You won't be able to undo this",
  confirmLabel = "Delete",
  confirmingLabel = "Deleting...",
  cancelLabel = "Keep",
}: Props) {
  const [isDeleting, setIsDeleting] = useState(false);

  async function handleConfirm() {
    setIsDeleting(true);
    try {
      await onConfirm();
      onOpenChange(false);
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <Dialog.Root
      open={isOpen}
      onOpenChange={(e) => onOpenChange(e.open)}
      lazyMount
      unmountOnExit
    >
      <Portal>
        <Dialog.Backdrop className="deletesheet__backdrop" />
        <Dialog.Positioner className="deletesheet__positioner">
          <Dialog.Content className="deletesheet__content">
            <VStack
              w="full"
              minWidth={{ base: "full", md: "sm" }}
              maxWidth={{ base: "full", md: "md" }}
              alignItems="stretch"
              gap="0"
              bg="bg.surfaceWhite"
              borderTopRadius="[28px]"
            >
              <styled.div display="flex" justifyContent="center" pt="3" pb="2">
                <styled.div
                  w="10"
                  h="[5px]"
                  borderRadius="[100px]"
                  bg="border.default"
                />
              </styled.div>

              <VStack alignItems="stretch" gap="6" px="5" pt="3" pb="10">
                <VStack alignItems="center" gap="1">
                  <Dialog.Title asChild>
                    <styled.p
                      m="0"
                      fontSize="lg"
                      lineHeight="[22px]"
                      fontWeight="semibold"
                      color="fg.default"
                      textAlign="center"
                    >
                      {title}
                    </styled.p>
                  </Dialog.Title>
                  <Dialog.Description asChild>
                    <styled.p
                      m="0"
                      fontSize="xs"
                      lineHeight="[16px]"
                      color="fg.muted"
                      textAlign="center"
                    >
                      {message}
                    </styled.p>
                  </Dialog.Description>
                </VStack>

                <styled.div display="flex" alignItems="center" gap="3" w="full">
                  <styled.button
                    type="button"
                    onClick={handleConfirm}
                    disabled={isDeleting}
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    flex="1"
                    h="12"
                    borderRadius="[12px]"
                    fontSize="md"
                    lineHeight="[20px]"
                    fontWeight="semibold"
                    color="fg.default"
                    bg="bg.profileStats"
                    border="none"
                    cursor={isDeleting ? "default" : "pointer"}
                    _disabled={{ opacity: "[0.7]" }}
                  >
                    {isDeleting ? confirmingLabel : confirmLabel}
                  </styled.button>

                  <Dialog.CloseTrigger asChild>
                    <styled.button
                      type="button"
                      disabled={isDeleting}
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      flex="1"
                      h="12"
                      borderRadius="[12px]"
                      fontSize="md"
                      lineHeight="[20px]"
                      fontWeight="semibold"
                      color="white"
                      bg="bg.composerSubmit"
                      border="none"
                      cursor="pointer"
                    >
                      {cancelLabel}
                    </styled.button>
                  </Dialog.CloseTrigger>
                </styled.div>
              </VStack>
            </VStack>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>

      <style jsx global>{`
        .deletesheet__backdrop {
          position: fixed;
          inset: 0;
          background-color: rgba(0, 0, 0, 0.5);
          z-index: var(--z-index-overlay);
        }
        .deletesheet__positioner {
          position: fixed;
          inset: 0;
          z-index: var(--z-index-modal);
          display: flex;
          flex-direction: column;
          justify-content: flex-end;
          align-items: center;
        }
        @media screen and (min-width: 48em) {
          .deletesheet__positioner {
            justify-content: center;
          }
        }
        .deletesheet__content {
          width: 100%;
          display: flex;
          justify-content: center;
        }
      `}</style>
    </Dialog.Root>
  );
}
