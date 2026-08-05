"use client";

import { Dialog, Portal } from "@ark-ui/react";
import { Camera, FolderClosed, Images } from "lucide-react";
import { ReactNode } from "react";

import { VStack, styled } from "@/styled-system/jsx";

type Props = {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  /**
   * Only phones and tablets have a photo library distinct from the filesystem.
   * On a laptop it would open the same browser as "Choose from Files", so the
   * entry is hidden there rather than duplicating it.
   */
  showGallery?: boolean;
  /** Photo library — images only. */
  onUploadGallery: () => void;
  /** Any file the OS browser offers, not just images. */
  onChooseFromFiles: () => void;
  /** Opens the camera directly on devices that have one. */
  onTakePhoto: () => void;
};

export function MediaUploadSheet({
  isOpen,
  onOpenChange,
  showGallery = true,
  onUploadGallery,
  onChooseFromFiles,
  onTakePhoto,
}: Props) {
  function pick(action: () => void) {
    // Close first: the file dialog is a native window, and leaving the sheet
    // mounted behind it means it is still there when the user cancels.
    onOpenChange(false);
    action();
  }

  return (
    <Dialog.Root
      open={isOpen}
      onOpenChange={(e) => onOpenChange(e.open)}
      lazyMount
      unmountOnExit
    >
      <Portal>
        <Dialog.Backdrop className="mediasheet__backdrop" />
        <Dialog.Positioner className="mediasheet__positioner">
          <Dialog.Content className="mediasheet__content">
            <VStack
              w="full"
              minWidth={{ base: "full", md: "sm" }}
              maxWidth={{ base: "full", md: "md" }}
              gap="0"
              bg="bg.surfaceWhite"
              style={{
                borderTopLeftRadius: "28px",
                borderTopRightRadius: "28px",
              }}
            >
              {/* Home indicator */}
              <styled.div display="flex" justifyContent="center" w="full" py="[11px]">
                <styled.div
                  style={{
                    width: "104px",
                    height: "5px",
                    borderRadius: "100px",
                    backgroundColor: "#DEDEDE",
                  }}
                />
              </styled.div>

              <VStack w="full" gap="[12px]" pt="[20px]" pb="[40px]" px="[20px]">
                {showGallery && (
                  <SheetButton
                    icon={<Images size={20} />}
                    label="Upload Gallery"
                    onClick={() => pick(onUploadGallery)}
                  />
                )}
                <SheetButton
                  icon={<FolderClosed size={20} />}
                  label="Choose from Files"
                  onClick={() => pick(onChooseFromFiles)}
                />
                <SheetButton
                  icon={<Camera size={20} />}
                  label="Take A Photo"
                  onClick={() => pick(onTakePhoto)}
                  variant="solid"
                />
              </VStack>
            </VStack>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>

      <style jsx global>{`
        .mediasheet__backdrop {
          position: fixed;
          inset: 0;
          background-color: rgba(0, 0, 0, 0.5);
          z-index: var(--z-index-overlay);
        }
        .mediasheet__positioner {
          position: fixed;
          inset: 0;
          z-index: var(--z-index-modal);
          display: flex;
          flex-direction: column;
          justify-content: flex-end;
          align-items: center;
        }
        @media screen and (min-width: 48em) {
          .mediasheet__positioner {
            justify-content: center;
          }
        }
        .mediasheet__content {
          width: 100%;
          display: flex;
          justify-content: center;
        }
      `}</style>
    </Dialog.Root>
  );
}

function SheetButton({
  icon,
  label,
  onClick,
  variant = "subtle",
}: {
  icon: ReactNode;
  label: string;
  onClick: () => void;
  variant?: "subtle" | "solid";
}) {
  const isSolid = variant === "solid";

  return (
    <styled.button
      type="button"
      onClick={onClick}
      display="flex"
      alignItems="center"
      justifyContent="center"
      gap="2"
      w="full"
      style={{
        height: "48px",
        paddingLeft: "16px",
        paddingRight: "20px",
        borderRadius: "12px",
        border: "none",
        cursor: "pointer",
        backgroundColor: isSolid ? "#404040" : "#F0F0F0",
        color: isSolid ? "#FFFFFF" : "#404040",
        fontSize: "16px",
        lineHeight: "20px",
        fontWeight: 600,
      }}
    >
      {icon}
      {label}
    </styled.button>
  );
}
