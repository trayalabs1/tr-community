import Image, { ImageOptions } from "@tiptap/extension-image";
import {
  NodeViewProps,
  NodeViewWrapper,
  ReactNodeViewRenderer,
  mergeAttributes,
} from "@tiptap/react";
import { Plugin, PluginKey } from "prosemirror-state";
import { EditorView } from "prosemirror-view";
import { Trash2 } from "lucide-react";

import { Asset } from "src/api/openapi-schema";

import { Button } from "@/components/ui/button";
import { ProgressCircle } from "@/components/ui/progress";
import { css } from "@/styled-system/css";
import { styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

// NOTE: This is the name of the component that will be used in the HTML.
// It cannot be changed.
const COMPONENT_NAME = "img";

// NOTE: This plugin key is used to store the upload positions in ProseMirror
// state to remove the need to scan the whole document for positions.
export const uploadPositionsKey = new PluginKey<Map<string, number>>(
  "uploadPositions",
);

type Options = {
  handleFiles: (view: EditorView, files: File[]) => Promise<Asset[]>;
  handleRetry: (view: EditorView, uploadId: string) => void;
  handleCancel: (view: EditorView, uploadId: string) => void;
};

function Component(props: NodeViewProps) {
  const isUploading = props.node.attrs["data-uploading"] === "true";
  const uploadError = props.node.attrs["data-upload-error"];
  const uploadId = props.node.attrs["data-upload-id"];
  const uploadProgress = props.node.attrs["data-upload-progress"];
  const progressPercent = uploadProgress ? parseInt(uploadProgress, 10) : 0;

  const isEditable = props.editor.isEditable;
  const isSelected = props.selected && isEditable;

  const { handleRetry, handleCancel } = props.extension.options as Options;

  return (
    <NodeViewWrapper
      className={css({
        display: "block",
        cursor: "pointer",
      })}
      style={{ width: "fit-content" }}
    >
      <styled.div
        position="relative"
        display="inline-block"
        borderRadius="lg"
        userSelect={isEditable ? "none" : "auto"}
        style={{
          width: "180px",
          height: "auto",
        }}
        onMouseEnter={(e) => {
          const deleteBtn = (e.currentTarget as HTMLElement).querySelector(
            ".image-delete-btn"
          ) as HTMLElement;
          if (deleteBtn) {
            deleteBtn.style.opacity = "1";
          }
        }}
        onMouseLeave={(e) => {
          const deleteBtn = (e.currentTarget as HTMLElement).querySelector(
            ".image-delete-btn"
          ) as HTMLElement;
          if (deleteBtn) {
            deleteBtn.style.opacity = "0";
          }
        }}
      >
        <styled.img
          borderRadius="lg"
          opacity={isUploading ? "5" : "full"}
          transition="all"
          w="full"
          h="auto"
          style={{
            display: "block",
          }}
          {...props.node.attrs}
        />
        {isUploading && (
          <styled.div
            position="absolute"
            top="0"
            left="0"
            width="full"
            height="full"
            display="flex"
            flexDirection="column"
            alignItems="center"
            justifyContent="center"
            pointerEvents="none"
            gap="3"
            padding="4"
            borderRadius="lg"
            style={{
              backgroundColor: "rgba(255, 255, 255, 0.8)",
            }}
          >
            <ProgressCircle value={progressPercent} size="md" />
          </styled.div>
        )}
        {uploadError && (
          <styled.div
            position="absolute"
            inset="0"
            display="flex"
            flexDirection="column"
            alignItems="center"
            justifyContent="center"
            backgroundColor="bg.error"
            opacity="9"
            borderRadius="md"
            padding="3"
            gap="2"
            userSelect="none"
            contentEditable={false}
          >
            <styled.p fontSize="sm" color="fg.error" fontWeight="medium">
              Upload failed
            </styled.p>
            <styled.div display="flex" gap="2">
              <Button
                type="button"
                size="xs"
                variant="outline"
                onClick={() => handleRetry(props.view, uploadId)}
              >
                Retry
              </Button>
              <Button
                type="button"
                size="xs"
                variant="ghost"
                onClick={() => handleCancel(props.view, uploadId)}
              >
                Remove
              </Button>
            </styled.div>
          </styled.div>
        )}
        {!isUploading && !uploadError && isEditable && (
          <styled.button
            type="button"
            className="image-delete-btn"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              try {
                if (uploadId) {
                  handleCancel(props.view, uploadId);
                } else {
                  const pos = props.getPos();
                  const tr = props.view.state.tr;
                  tr.deleteRange(pos, pos + props.node.nodeSize);
                  props.view.dispatch(tr);
                }
              } catch (err) {
                console.error("Error deleting image:", err);
              }
            }}
            display="flex"
            alignItems="center"
            justifyContent="center"
            position="absolute"
            style={{
              top: "0.5rem",
              right: "0.5rem",
              width: "32px",
              height: "32px",
              backgroundColor: TRAYA_COLORS.heart,
              border: "none",
              borderRadius: "0.375rem",
              cursor: "pointer",
              color: "white",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              opacity: "0",
              transition: "opacity 0.2s ease-in-out",
              zIndex: 10,
              boxShadow: "0 2px 8px rgba(0, 0, 0, 0.15)",
              pointerEvents: "auto",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor =
                TRAYA_COLORS.fallback;
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor =
                TRAYA_COLORS.heart;
            }}
            title="Delete image"
          >
            <Trash2 size={18} strokeWidth={2} />
          </styled.button>
        )}
      </styled.div>
    </NodeViewWrapper>
  );
}

export const ImageExtended = Image.extend<ImageOptions & Options>({
  content: "inline*",
  addOptions() {
    return {
      ...this.parent?.(),
    };
  },
  addAttributes() {
    return {
      ...this.parent?.(),
      "data-upload-id": {
        default: null,
        parseHTML: (element) => element.getAttribute("data-upload-id"),
        renderHTML: (attributes) => {
          if (!attributes["data-upload-id"]) {
            return {};
          }
          return {
            "data-upload-id": attributes["data-upload-id"],
          };
        },
      },
      "data-uploading": {
        default: null,
        parseHTML: (element) => element.getAttribute("data-uploading"),
        renderHTML: (attributes) => {
          if (!attributes["data-uploading"]) {
            return {};
          }
          return {
            "data-uploading": attributes["data-uploading"],
          };
        },
      },
      "data-upload-error": {
        default: null,
        parseHTML: (element) => element.getAttribute("data-upload-error"),
        renderHTML: (attributes) => {
          if (!attributes["data-upload-error"]) {
            return {};
          }
          return {
            "data-upload-error": attributes["data-upload-error"],
          };
        },
      },
      "data-upload-progress": {
        default: null,
        parseHTML: (element) => element.getAttribute("data-upload-progress"),
        renderHTML: (attributes) => {
          if (!attributes["data-upload-progress"]) {
            return {};
          }
          return {
            "data-upload-progress": attributes["data-upload-progress"],
          };
        },
      },
    };
  },
  addNodeView() {
    return ReactNodeViewRenderer(Component);
  },
  parseHTML() {
    return [
      {
        tag: COMPONENT_NAME,
      },
    ];
  },
  renderHTML({ HTMLAttributes }) {
    return [COMPONENT_NAME, mergeAttributes(HTMLAttributes), 0];
  },
  addProseMirrorPlugins() {
    const handleFiles = this.options.handleFiles;
    return [
      // Position tracking plugin - maintains a map of uploadId -> position
      new Plugin({
        key: uploadPositionsKey,
        state: {
          init() {
            return new Map<string, number>();
          },
          apply(_tr, _oldMap, _oldState, newState) {
            // Rebuild position map by scanning the document
            const newMap = new Map<string, number>();

            newState.doc.descendants((node, pos) => {
              const uploadId = node.attrs["data-upload-id"];

              if (uploadId) {
                newMap.set(uploadId, pos);
              }
            });

            return newMap;
          },
        },
      }),

      // Paste handler plugin
      new Plugin({
        props: {
          handlePaste(view, event) {
            if (!event.clipboardData) {
              return false;
            }

            const files: File[] = [];

            // Use "items"
            if (event.clipboardData.items?.length) {
              for (const item of event.clipboardData.items) {
                if (item.kind === "file") {
                  const file = item.getAsFile();
                  if (file) {
                    files.push(file);
                  }
                }
              }
            }

            const images = files.filter((file) => /image/i.test(file.type));

            if (images.length === 0) {
              return false;
            }

            event.preventDefault();
            handleFiles(view, images);
            return true;
          },
        },
      }),
    ];
  },
});
