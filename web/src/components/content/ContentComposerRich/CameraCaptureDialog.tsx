"use client";

import { Dialog, Portal } from "@ark-ui/react";
import { useCallback, useEffect, useRef, useState } from "react";

import { VStack, styled } from "@/styled-system/jsx";

type Props = {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onCapture: (file: File) => void;
};

// Desktop browsers ignore <input capture>, so taking a photo there needs a real
// camera stream. Mobile keeps using the input, which opens the native camera.
export function CameraCaptureDialog({ isOpen, onOpenChange, onCapture }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isReady, setReady] = useState(false);

  const stop = useCallback(() => {
    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;
    setReady(false);
  }, []);

  useEffect(() => {
    if (!isOpen) {
      stop();
      return;
    }

    let cancelled = false;
    setError(null);

    navigator.mediaDevices
      ?.getUserMedia({ video: { facingMode: "environment" }, audio: false })
      .then((stream) => {
        // The dialog can close while getUserMedia is still resolving; without
        // this the camera light stays on with no way to switch it off.
        if (cancelled) {
          stream.getTracks().forEach((t) => t.stop());
          return;
        }
        streamRef.current = stream;
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
        }
        setReady(true);
      })
      .catch(() => {
        if (!cancelled) {
          setError(
            "Camera unavailable. Check the browser's camera permission and try again.",
          );
        }
      });

    return () => {
      cancelled = true;
      stop();
    };
  }, [isOpen, stop]);

  function handleShutter() {
    const video = videoRef.current;
    if (!video) return;

    const canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext("2d")?.drawImage(video, 0, 0);

    canvas.toBlob((blob) => {
      if (!blob) return;
      onCapture(
        new File([blob], `photo-${canvas.width}x${canvas.height}.jpg`, {
          type: "image/jpeg",
        }),
      );
      onOpenChange(false);
    }, "image/jpeg", 0.92);
  }

  return (
    <Dialog.Root
      open={isOpen}
      onOpenChange={(e) => onOpenChange(e.open)}
      lazyMount
      unmountOnExit
    >
      <Portal>
        <Dialog.Backdrop className="camera__backdrop" />
        <Dialog.Positioner className="camera__positioner">
          <Dialog.Content className="camera__content">
            <VStack
              w="full"
              maxWidth={{ base: "full", md: "lg" }}
              gap="0"
              bg="bg.surfaceWhite"
              overflow="hidden"
              style={{ borderRadius: "16px" }}
            >
              {error ? (
                <styled.p p="6" textAlign="center" style={{ color: "#404040" }}>
                  {error}
                </styled.p>
              ) : (
                <styled.video
                  ref={videoRef}
                  autoPlay
                  playsInline
                  muted
                  w="full"
                  style={{ aspectRatio: "1 / 1", objectFit: "cover", backgroundColor: "#000000" }}
                />
              )}

              <styled.div display="flex" gap="3" w="full" p="4">
                <styled.button
                  type="button"
                  onClick={() => onOpenChange(false)}
                  flex="1"
                  style={{
                    height: "48px",
                    borderRadius: "12px",
                    border: "none",
                    cursor: "pointer",
                    backgroundColor: "#F0F0F0",
                    color: "#404040",
                    fontSize: "16px",
                    fontWeight: 600,
                  }}
                >
                  Cancel
                </styled.button>
                {!error && (
                  <styled.button
                    type="button"
                    onClick={handleShutter}
                    disabled={!isReady}
                    flex="1"
                    style={{
                      height: "48px",
                      borderRadius: "12px",
                      border: "none",
                      cursor: isReady ? "pointer" : "not-allowed",
                      backgroundColor: isReady ? "#404040" : "#C9C9C9",
                      color: "#FFFFFF",
                      fontSize: "16px",
                      fontWeight: 600,
                    }}
                  >
                    Capture
                  </styled.button>
                )}
              </styled.div>
            </VStack>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>

      <style jsx global>{`
        .camera__backdrop {
          position: fixed;
          inset: 0;
          background-color: rgba(0, 0, 0, 0.5);
          z-index: var(--z-index-overlay);
        }
        .camera__positioner {
          position: fixed;
          inset: 0;
          z-index: var(--z-index-modal);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 16px;
        }
        .camera__content {
          width: 100%;
          display: flex;
          justify-content: center;
        }
      `}</style>
    </Dialog.Root>
  );
}
