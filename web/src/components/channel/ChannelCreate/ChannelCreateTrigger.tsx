"use client";

import { Button, ButtonProps } from "@/components/ui/button";
import { CreateIcon } from "@/components/ui/icons/Create";
import { useDisclosure } from "@/utils/useDisclosure";

import { ChannelCreateModal } from "./ChannelCreateModal";

type Props = ButtonProps & {
  hideLabel?: boolean;
};

export function ChannelCreateTrigger({ hideLabel, ...props }: Props) {
  const useDisclosureProps = useDisclosure();

  return (
    <>
      <Button {...props} onClick={useDisclosureProps.onOpen}>
        <CreateIcon />
        {!hideLabel && <span>Create Channel</span>}
      </Button>

      <ChannelCreateModal {...useDisclosureProps} />
    </>
  );
}
