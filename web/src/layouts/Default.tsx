import { PropsWithChildren, ReactNode } from "react";

import { Navigation } from "src/components/site/Navigation/Navigation";

import { Box, styled } from "@/styled-system/jsx";

type Props = {
  contextpane: ReactNode;
};

export async function Default({
  contextpane,
  children,
}: PropsWithChildren<Props>) {
  return (
    <Navigation contextpane={contextpane} paddingMode="none">
      <styled.main
        containerType="inline-size"
        width="full"
        height="full"
        minW="0"
      >
        {children}
        <Box height="24"></Box>
      </styled.main>
    </Navigation>
  );
}
