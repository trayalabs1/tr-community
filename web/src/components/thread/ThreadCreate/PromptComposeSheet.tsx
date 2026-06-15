"use client";

import { Drawer } from "vaul";
import { useEffect, useState } from "react";

import { handle } from "@/api/client";
import { channelThreadCreate } from "@/api/openapi-client/channels";
import { Account, Permission, Visibility } from "@/api/openapi-schema";
import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { MemberAvatar } from "@/components/member/MemberBadge/MemberAvatar";
import { Button } from "@/components/ui/button";
import { CloseAction } from "@/components/site/Action/Close";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { HStack, LStack, VStack, WStack, styled } from "@/styled-system/jsx";
import { hasPermission } from "@/utils/permissions";
import { OpenChangeEvent } from "@/utils/useDisclosure";
import { useFeedMutations } from "@/lib/feed/mutation";

type Props = {
  isOpen: boolean;
  onClose: () => void;
  onOpenChange: (event: OpenChangeEvent) => void;
  channelID: string;
  channelName?: string;
  session?: Account;
  initialText?: string;
  tag?: string;
  tagIndex?: number;
};

const ACCENT = "#329866";
const ACCENT_TINT = "#DCF4E8";

export function PromptComposeSheet({
  isOpen,
  onOpenChange,
  channelID,
  channelName,
  session,
  initialText,
  tag,
  tagIndex,
}: Props) {
  const [body, setBody] = useState("");
  const [resetKey, setResetKey] = useState("");
  const [isPosting, setIsPosting] = useState(false);
  const { trackSubmitForReview } = useEventTracking();
  const { revalidate } = useFeedMutations(session, undefined, channelID);

  useEffect(() => {
    if (isOpen) {
      setBody("");
      setResetKey(`prompt-${initialText ?? ""}-${isOpen}`);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, initialText]);

  function handleOpenChange(open: boolean) {
    if (!open) {
      setBody("");
    }
    onOpenChange({ open });
  }

  function bodyText(html: string): string {
    return (
      new DOMParser()
        .parseFromString(html, "text/html")
        .querySelector("body")?.textContent?.trim() ?? ""
    );
  }

  async function handlePost() {
    if (bodyText(body) === "") {
      return;
    }

    const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);
    const visibility = isAdmin ? Visibility.published : Visibility.review;

    await handle(
      async () => {
        setIsPosting(true);
        trackSubmitForReview(body.length, false, false, channelID);

        await channelThreadCreate(channelID, {
          title: "",
          body,
          visibility,
        });

        await revalidate();
        handleOpenChange(false);
      },
      {
        promiseToast: {
          loading: "Publishing post...",
          success: channelName ? `Posted to ${channelName}` : "Post published!",
        },
        cleanup: async () => {
          setIsPosting(false);
        },
      },
    );
  }

  return (
    <Drawer.Root
      open={isOpen}
      onOpenChange={handleOpenChange}
      shouldScaleBackground={false}
      repositionInputs={false}
    >
      <Drawer.Portal>
        <Drawer.Overlay className="promptsheet__overlay" />
        <Drawer.Content className="promptsheet__content">
          <VStack
            w="full"
            minWidth={{ base: "full", md: "lg" }}
            maxWidth={{ base: "full", md: "xl" }}
            borderTopRadius="2xl"
            boxShadow="xl"
            p="4"
            gap="4"
            alignItems="stretch"
            style={{
              backgroundColor: "#ffffff",
              maxHeight: "var(--modal-vh, 90dvh)",
            }}
          >
            <WStack alignItems="start">
              <HStack gap="3" alignItems="center">
                {session && <MemberAvatar profile={session} size="sm" />}
                <LStack gap="0">
                  <styled.span fontSize="sm" fontWeight="bold">
                    Posting as @{session?.handle}
                  </styled.span>
                  {channelName && (
                    <styled.span fontSize="xs" color="fg.muted">
                      to {channelName}
                    </styled.span>
                  )}
                </LStack>
              </HStack>
              <CloseAction onClick={() => handleOpenChange(false)} />
            </WStack>

            <styled.div
              w="full"
              flex="1"
              fontSize="xl"
              style={{ minHeight: "5.5rem" }}
            >
              <ContentComposer
                key={resetKey}
                onChange={setBody}
                placeholder={initialText}
              />
            </styled.div>

            <WStack
              alignItems="center"
              pt="3"
              gap="2"
              style={{ borderTop: "1px solid var(--colors-border-default)" }}
            >
              <HStack gap="2" alignItems="center" flex="1" minW="0">
                {tag && (
                  <styled.span
                    flexShrink="0"
                    display="inline-flex"
                    alignItems="center"
                    gap="1"
                    px="2"
                    py="0.5"
                    rounded="full"
                    fontSize="2xs"
                    fontWeight="bold"
                    letterSpacing="wide"
                    textTransform="uppercase"
                    style={{ background: ACCENT_TINT, color: ACCENT }}
                  >
                    {tagIndex !== undefined && (
                      <styled.span style={{ opacity: 0.7 }}>
                        {String(tagIndex).padStart(2, "0")}
                      </styled.span>
                    )}
                    {tag}
                  </styled.span>
                )}
                <styled.span
                  fontSize="xs"
                  color="fg.muted"
                  whiteSpace="nowrap"
                  overflow="hidden"
                  textOverflow="ellipsis"
                >
                  Edit before posting
                </styled.span>
              </HStack>

              <Button
                type="button"
                size="sm"
                loading={isPosting}
                onClick={handlePost}
                flexShrink="0"
                style={{ backgroundColor: ACCENT, color: "white" }}
              >
                Post →
              </Button>
            </WStack>
          </VStack>
        </Drawer.Content>
      </Drawer.Portal>

      <style jsx global>{`
        .promptsheet__overlay {
          position: fixed;
          inset: 0;
          background-color: rgba(0, 0, 0, 0.5);
          z-index: 50;
        }
        .promptsheet__content {
          position: fixed;
          left: 0;
          right: 0;
          bottom: var(--modal-kb-offset, 0px);
          z-index: 51;
          display: flex;
          flex-direction: column;
          align-items: center;
          max-height: var(--modal-vh, 90dvh);
        }
        @media screen and (min-width: 48em) {
          .promptsheet__content {
            top: 0;
            bottom: 0;
            justify-content: center;
          }
        }
      `}</style>
    </Drawer.Root>
  );
}
