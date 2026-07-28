"use client";

import { Dialog, Portal } from "@ark-ui/react";
import { useState } from "react";

import { VStack, styled } from "@/styled-system/jsx";

type Props = {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void | Promise<void>;
  title?: string;
  message?: string;
};

const DESTRUCTIVE = "#E54336";
const MUTED = "#787878";
const BORDER = "#dedede";
const TEXT = "#404040";

export function DeleteConfirmSheet({
  isOpen,
  onOpenChange,
  onConfirm,
  title = "Delete this post?",
  message = "This action cannot be undone.",
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
              style={{
                backgroundColor: "#ffffff",
                borderTopLeftRadius: "28px",
                borderTopRightRadius: "28px",
              }}
            >
              <styled.div display="flex" justifyContent="center" pt="3" pb="2">
                <styled.div
                  style={{
                    width: "40px",
                    height: "5px",
                    borderRadius: "100px",
                    backgroundColor: BORDER,
                  }}
                />
              </styled.div>

              <VStack alignItems="stretch" gap="5" style={{ padding: "20px 20px 40px" }}>
                <VStack alignItems="center" gap="2">
                  <Dialog.Title asChild>
                    <styled.p
                      fontWeight="semibold"
                      style={{ margin: 0, fontSize: "16px", lineHeight: "20px", color: TEXT }}
                    >
                      {title}
                    </styled.p>
                  </Dialog.Title>
                  <Dialog.Description asChild>
                    <styled.p
                      fontWeight="medium"
                      style={{
                        margin: 0,
                        fontSize: "14px",
                        lineHeight: "20px",
                        color: MUTED,
                        textAlign: "center",
                      }}
                    >
                      {message}
                    </styled.p>
                  </Dialog.Description>
                </VStack>

                <VStack alignItems="stretch" gap="3">
                  <styled.button
                    type="button"
                    onClick={handleConfirm}
                    disabled={isDeleting}
                    display="flex"
                    alignItems="center"
                    justifyContent="center"
                    w="full"
                    style={{
                      height: "48px",
                      borderRadius: "12px",
                      backgroundColor: DESTRUCTIVE,
                      color: "#ffffff",
                      fontSize: "16px",
                      fontWeight: 600,
                      lineHeight: "20px",
                      border: "none",
                      cursor: isDeleting ? "default" : "pointer",
                      opacity: isDeleting ? 0.7 : 1,
                    }}
                  >
                    {isDeleting ? "Deleting..." : "Delete post"}
                  </styled.button>

                  <Dialog.CloseTrigger asChild>
                    <styled.button
                      type="button"
                      disabled={isDeleting}
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      w="full"
                      style={{
                        height: "48px",
                        borderRadius: "12px",
                        backgroundColor: "#ffffff",
                        color: TEXT,
                        fontSize: "16px",
                        fontWeight: 600,
                        lineHeight: "20px",
                        border: `1px solid ${BORDER}`,
                        cursor: "pointer",
                      }}
                    >
                      Cancel
                    </styled.button>
                  </Dialog.CloseTrigger>
                </VStack>
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
          z-index: 50;
        }
        .deletesheet__positioner {
          position: fixed;
          inset: 0;
          z-index: 51;
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
