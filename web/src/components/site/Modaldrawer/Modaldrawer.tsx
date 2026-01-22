import { PropsWithChildren } from "react";
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

  return (
    <>
      <Drawer.Root
        open={props.isOpen}
        onOpenChange={handleOpenChange}
        // TODO: Scale background only on mobile.
        shouldScaleBackground={false}
        dismissible={props.dismissable}
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
                maxHeight: "90vh",
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
          background-color: var(--colors-black-alpha-600);
          z-index: var(--z-index-overlay);
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
            z-index: var(--z-index-modal);
          }
        }

        /* Drawer mode - on mobile screens */
        @media screen and (max-width: 48em) {
          .modaldrawer__content {
            height: auto;
            width: 100%;
            bottom: 0;
            top: auto;
            display: flex;
            flex-direction: column;
            position: fixed;
            max-height: 90vh;
            z-index: var(--z-index-modal);
            border-radius: 1.5rem 1.5rem 0 0;
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
