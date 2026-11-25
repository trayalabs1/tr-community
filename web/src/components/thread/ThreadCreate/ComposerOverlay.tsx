"use client";

import { Dialog } from "@ark-ui/react";
import { PropsWithChildren, useEffect } from "react";

import { useSession } from "@/auth";
import { SelfAvatarBadge } from "@/components/member/MemberBadge/SelfAvatarBadge";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { NotificationIcon } from "@/components/ui/icons/Notification";
import { SearchIcon } from "@/components/ui/icons/Search";
import { HStack, styled } from "@/styled-system/jsx";
import { UseDisclosureProps } from "@/utils/useDisclosure";

type Props = {
  title?: string;
} & UseDisclosureProps;

export function ComposerOverlay({
  children,
  title,
  ...props
}: PropsWithChildren<Props>) {
  const session = useSession();

  useEffect(() => {
    if (!props.isOpen || typeof window === "undefined" || !window.visualViewport)
      return;

    const vv = window.visualViewport;

    const update = () => {
      document.documentElement.style.setProperty(
        "--composer-vh",
        `${vv.height}px`,
      );
    };

    vv.addEventListener("resize", update);
    vv.addEventListener("scroll", update);
    update();

    return () => {
      vv.removeEventListener("resize", update);
      vv.removeEventListener("scroll", update);
      document.documentElement.style.removeProperty("--composer-vh");
    };
  }, [props.isOpen]);

  return (
    <Dialog.Root
      open={props.isOpen}
      onOpenChange={({ open }) => {
        if (open) {
          props.onOpen?.();
        } else {
          props.onClose?.();
        }
        props.onOpenChange?.({ open });
      }}
      unmountOnExit
      lazyMount
    >
      <Dialog.Backdrop className="composer__backdrop" />
      <Dialog.Positioner className="composer__positioner">
        <Dialog.Content className="composer__content">
          <styled.div
            display="flex"
            flexDirection="column"
            w="full"
            h="full"
            minH="0"
            maxH="full"
            overflow="hidden"
            bg="bg.composerPage"
          >
            <styled.div
              display="flex"
              alignItems="center"
              justifyContent="space-between"
              gap="2"
              flexShrink="0"
              w="full"
              h="14"
              pl="1"
              pr="3.5"
              bg="bg.searchHeader"
            >
              <styled.button
                type="button"
                aria-label="Back"
                onClick={props.onClose}
                display="flex"
                alignItems="center"
                justifyContent="center"
                w="9"
                h="9"
                p="0"
                flexShrink="0"
                color="fg.default"
                bg="transparent"
                border="none"
                cursor="pointer"
              >
                <ArrowLeftIcon width="6" height="6" />
              </styled.button>

              <Dialog.Title asChild>
                <styled.span srOnly>{title}</styled.span>
              </Dialog.Title>

              <HStack gap="2" alignItems="center" flexShrink="0">
                <styled.span
                  display="inline-flex"
                  alignItems="center"
                  justifyContent="center"
                  w="10"
                  h="10"
                  color="fg.default"
                >
                  <NotificationIcon width="5" height="5" />
                </styled.span>

                <styled.span
                  display="inline-flex"
                  alignItems="center"
                  justifyContent="center"
                  w="10"
                  h="10"
                  color="fg.default"
                >
                  <SearchIcon width="5" height="5" />
                </styled.span>

                {session && (
                  <SelfAvatarBadge account={session} asLink={false} />
                )}
              </HStack>
            </styled.div>

            <styled.div
              flex="1"
              minH="0"
              w="full"
              overflowY="auto"
              overflowX="hidden"
            >
              <styled.div w="full" maxW="2xl" mx="auto">
                {children}
              </styled.div>
            </styled.div>
          </styled.div>
        </Dialog.Content>
      </Dialog.Positioner>

      <style jsx global>{`
        .composer__backdrop {
          position: fixed;
          inset: 0;
          background-color: rgba(0, 0, 0, 0.5);
          z-index: 50;
        }
        .composer__positioner {
          position: fixed;
          inset: 0;
          z-index: 51;
          display: flex;
        }
        .composer__content {
          display: flex;
          flex-direction: column;
          width: 100%;
          height: var(--composer-vh, 100dvh);
          max-height: 100dvh;
          overflow: hidden;
        }
        @media screen and (min-width: 48em) {
          .composer__positioner {
            align-items: center;
            justify-content: center;
            padding: 1.5rem;
          }
          .composer__content {
            width: 100%;
            max-width: 56rem;
            height: 85dvh;
            max-height: 85dvh;
            border-radius: 1rem;
          }
        }
      `}</style>
    </Dialog.Root>
  );
}
