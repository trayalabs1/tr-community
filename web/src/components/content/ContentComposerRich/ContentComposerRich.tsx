import { BubbleMenu, EditorContent } from "@tiptap/react";
import { ChangeEvent, useEffect, useRef, useState } from "react";

import { EditIcon } from "@/components/ui/icons/Edit";
import { MediaStackIcon } from "@/components/ui/icons/Media";
import { css, cx } from "@/styled-system/css";
import { HStack, LStack, styled } from "@/styled-system/jsx";

import { ComposerTools } from "../ComposerTools";
import { ContentDragOverlay } from "../ContentDragOverlay";
import { ContentComposerProps } from "../composer-props";

import "./styles.css";

import { CameraCaptureDialog } from "./CameraCaptureDialog";
import { EditorMenu } from "./EditorMenu";
import { LinkPasteMenu } from "./LinkPasteMenu";
import { MediaUploadSheet } from "./MediaUploadSheet";
import { useContentComposer } from "./useContentComposerRich";

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    linkPreview: {
      setLinkPreview: (attributes: { href: string }) => ReturnType;
    };
  }
}

export function ContentComposerRich(props: ContentComposerProps) {
  const {
    editor,
    initialValueHTML,
    uniqueID,
    uploadingCount,
    isDragging,
    isDragError,
    getDragOverlayMessage,
    handlers,
    format,
  } = useContentComposer(props);

  const [isMediaSheetOpen, setMediaSheetOpen] = useState(false);
  const [isCameraOpen, setCameraOpen] = useState(false);
  // Resolved after mount: reading matchMedia during render would disagree with
  // the server-rendered markup and trip a hydration error.
  const [isTouchDevice, setTouchDevice] = useState(false);
  const galleryInputRef = useRef<HTMLInputElement>(null);
  const filesInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setTouchDevice(window.matchMedia("(pointer: coarse)").matches);
  }, []);

  // <input capture> opens the native camera on mobile but is ignored on
  // desktop, where it falls back to a file browser. So desktop gets a real
  // getUserMedia capture instead.
  function handleTakePhoto() {
    const supportsCaptureAttribute = "capture" in document.createElement("input");

    if (supportsCaptureAttribute && isTouchDevice) {
      cameraInputRef.current?.click();
      return;
    }

    setCameraOpen(true);
  }

  // Route a captured photo through the same input the file picker uses, so it
  // takes the existing upload path rather than a second one that could drift.
  // React's onChange is synthetic and does not fire for a programmatic
  // dispatch, so the handler is invoked directly with the input as target.
  function handleCameraCapture(file: File) {
    const input = cameraInputRef.current;
    if (!input) return;

    const transfer = new DataTransfer();
    transfer.items.add(file);
    input.files = transfer.files;

    handlers.handleFileUpload({
      currentTarget: input,
      target: input,
    } as unknown as ChangeEvent<HTMLInputElement>);
  }

  return (
    <LStack
      id={`rich-text-editor-${uniqueID}`}
      containerType="inline-size"
      className={cx("typography", props.className)}
      position="relative"
      w="full"
      h={props.hideTools ? "full" : undefined}
      flex={props.hideTools ? "1" : undefined}
      gap="1"
      minHeight="8"
      onDragOver={handlers.handleDragOver}
      onDragEnter={handlers.handleDragEnter}
      onDragLeave={handlers.handleDragLeave}
      onDrop={handlers.handleDrop}
    >
      <div
        id={`editor-content-${uniqueID}`}
        className={css({
          height: "full",
          width: "full",
          flex: "1",
          minHeight: "0",
          position: "relative",
        })}
        suppressHydrationWarning
      >
        {editor ? (
          <EditorContent editor={editor} />
        ) : (
          <div dangerouslySetInnerHTML={{ __html: initialValueHTML }} />
        )}
      </div>
      {editor && props.hideTools && (
        <HStack w="full" justifyContent="flex-end" flexShrink="0" mt="auto">
          <styled.button
            type="button"
            onClick={() => setMediaSheetOpen(true)}
            display="flex"
            alignItems="center"
            justifyContent="center"
            w="10"
            h="10"
            flexShrink="0"
            borderRadius="[10px]"
            bg="bg.composerAction"
            color="fg.composerAction"
            cursor="pointer"
            border="none"
            title="Add an image"
            aria-label="Add an image"
          >
            <MediaStackIcon width="5" height="5" />
          </styled.button>

          {/* One input per source: the OS picker is configured by the accept
              and capture attributes, so each entry needs its own element. */}
          <styled.input
            ref={galleryInputRef}
            type="file"
            multiple
            display="none"
            accept="image/*"
            onChange={handlers.handleFileUpload}
          />
          <styled.input
            ref={filesInputRef}
            type="file"
            multiple
            display="none"
            onChange={handlers.handleFileUpload}
          />
          <styled.input
            ref={cameraInputRef}
            type="file"
            display="none"
            accept="image/*"
            capture="environment"
            onChange={handlers.handleFileUpload}
          />

          <MediaUploadSheet
            isOpen={isMediaSheetOpen}
            onOpenChange={setMediaSheetOpen}
            showGallery={isTouchDevice}
            onUploadGallery={() => galleryInputRef.current?.click()}
            onChooseFromFiles={() => filesInputRef.current?.click()}
            onTakePhoto={handleTakePhoto}
          />

          <CameraCaptureDialog
            isOpen={isCameraOpen}
            onOpenChange={setCameraOpen}
            onCapture={handleCameraCapture}
          />
        </HStack>
      )}

      {editor && !props.hideTools && (
        <div style={{ position: "relative", width: "100%", flexShrink: 0 }}>
          <ComposerTools
            enabled={!props.disabled}
            icon={<EditIcon />}
            workingCount={uploadingCount}
          >
            <EditorMenu
              editor={editor}
              uniqueID={`${uniqueID}-toolbar`}
              format={format}
              handlers={handlers}
            />
          </ComposerTools>
        </div>
      )}
      {editor && (
        <BubbleMenu
          editor={editor}
          tippyOptions={{
            placement: "bottom-start",
            maxWidth: "100%",
            popperOptions: {
              modifiers: [
                {
                  name: "offset",
                  options: {
                    offset: [0, 4],
                  },
                },
                {
                  name: "flip",
                  options: {
                    fallbackPlacements: ["top-start"],
                    boundary: editor.view.dom,
                    padding: 8,
                  },
                },
                {
                  name: "preventOverflow",
                  options: {
                    boundary: editor.view.dom,
                    altAxis: true,
                    padding: {
                      top: 0,
                      right: 0,
                      bottom: -40,
                      left: 0,
                    },
                    rootBoundary: "viewport",
                    tether: false,
                  },
                },
              ],
            },
          }}
          className={css({
            zIndex: "popover",
            borderRadius: "md",
            display: "flex",
            flexWrap: "wrap",
            minW: "0",
            maxW: "full",
            gap: "1",
            padding: "1",
            backgroundColor: "bg.subtle/80",
            backdropBlur: "frosted",
            backdropFilter: "auto",
            boxShadow: "md",
          })}
        >
          <EditorMenu
            editor={editor}
            uniqueID={`${uniqueID}-menu`}
            format={format}
            handlers={handlers}
          />
        </BubbleMenu>
      )}
      {editor && <LinkPasteMenu editor={editor} />}
      {isDragging && (
        <ContentDragOverlay
          isError={isDragError}
          message={getDragOverlayMessage()}
        />
      )}
    </LStack>
  );
}
