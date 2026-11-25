import { PropsWithChildren, useEffect } from "react";
import { Drawer } from "vaul";

import { UseDisclosureProps } from "src/utils/useDisclosure";

import { Heading } from "@/components/ui/heading";
import { Box, HStack, VStack, WStack } from "@/styled-system/jsx";

import { CloseAction } from "../Action/Close";

type Props = {
  title?: string;
  dismissable?: boolean;
} & UseDisclosureProps;

export function ModalDrawer({ children, ...props }: PropsWithChildren<Props>) {
  const handleOpenChange = (open: boolean) => {
    try {
      if (open) {
        props.onOpen?.();
      } else {
        props.onClose?.();
      }
    } finally {
      props.onOpenChange?.({ open });
    }
  };

  useEffect(() => {
    if (!props.isOpen || typeof window === "undefined" || !window.visualViewport)
      return;

    const vv = window.visualViewport;

    const update = () => {
      const kbOffset = Math.max(
        0,
        window.innerHeight - vv!.height - vv!.offsetTop,
      );
      const isKbOpen = kbOffset > 100;
      const maxH = isKbOpen ? vv!.height : vv!.height * 0.9;
      document.documentElement.style.setProperty("--modal-vh", `${maxH}px`);
      document.documentElement.style.setProperty(
        "--modal-kb-offset",
        `${kbOffset}px`,
      );
    };

    vv.addEventListener("resize", update);
    vv.addEventListener("scroll", update);
    update();

    return () => {
      vv.removeEventListener("resize", update);
      vv.removeEventListener("scroll", update);
      document.documentElement.style.removeProperty("--modal-vh");
      document.documentElement.style.removeProperty("--modal-kb-offset");
    };
  }, [props.isOpen]);

  return (
    <>
      <Drawer.Root
        open={props.isOpen}
        onOpenChange={handleOpenChange}
        shouldScaleBackground={false}
        dismissible={props.dismissable}
        repositionInputs={false}
      >
        <Drawer.Portal>
          <Drawer.Overlay className="modaldrawer__overlay" />
          <Drawer.Content className="modaldrawer__content">
            <VStack
              minHeight={{ base: "full", md: "0" }}
              minWidth={{ base: "full", md: "2xl" }}
              maxWidth={{ base: "full", md: "4xl" }}
              borderTopRadius={{ base: "2xl", md: "md" }}
              borderBottomRadius={{ base: "none", md: "md" }}
              boxShadow="xl"
              p={{ base: "3", md: "4" }}
              style={{
                backgroundColor: "#ffffff",
                maxHeight: "var(--modal-vh, 90dvh)",
              }}
            >
              <WStack alignItems="start">
                <Drawer.Title asChild>
                  <Heading size="md">{props.title}</Heading>
                </Drawer.Title>
                <CloseAction onClick={props.onClose} />
              </WStack>

              <Box h="full" w="full" pb="3" style={{ overflowY: "auto", overflowX: "hidden", touchAction: "auto", WebkitOverflowScrolling: "touch", minHeight: 0 }}>
                {children}
              </Box>
            </VStack>
          </Drawer.Content>
        </Drawer.Portal>
      </Drawer.Root>

      <style jsx global>{`
        .modaldrawer__overlay {
          position: fixed;
          inset: 0;
          background-color: rgba(0, 0, 0, 0.5);
          z-index: 50;
        }

        /* Modal mode - on desktop screens */
        @media screen and (min-width: 48em) {
          .modaldrawer__content {
            height: 100%;
            width: 100%;
            top: 0;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            position: fixed;
            z-index: 51;
          }
        }

        /* Drawer mode - on mobile screens */
        @media screen and (max-width: 48em) {
          .modaldrawer__content {
            height: auto;
            width: 100%;
            bottom: var(--modal-kb-offset, 0px);
            top: auto;
            display: flex;
            flex-direction: column;
            position: fixed;
            max-height: var(--modal-vh, 90dvh);
            z-index: 51;
            border-radius: 1rem 1rem 0 0;
            overflow-y: hidden;
            overflow-x: hidden;
            touch-action: none;
            scrollbar-width: none;
            -ms-overflow-style: none;
          }
          .modaldrawer__content::-webkit-scrollbar {
            display: none;
          }
        }
      `}</style>
    </>
  );
}
