import { useDisclosure } from "src/utils/useDisclosure";

import { ButtonProps } from "@/components/ui/button";
import { Button } from "@/components/ui/button";
import { AddIcon } from "@/components/ui/icons/Add";

import { ThreadCreateModal } from "./ThreadCreateModal";

type Props = ButtonProps & {
  channelID: string;
};

export function ThreadCreateTrigger({ channelID, ...props }: Props) {
  const useDisclosureProps = useDisclosure();

  return (
    <>
      <Button
        size="sm"
        onClick={useDisclosureProps.onOpen}
        {...props}
      >
        <AddIcon width="4" height="4" />
        New Thread
      </Button>

      <ThreadCreateModal
        {...useDisclosureProps}
        channelID={channelID}
      />
    </>
  );
}
