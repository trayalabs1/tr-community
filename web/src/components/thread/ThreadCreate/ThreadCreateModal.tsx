import { ChannelListOKResponse } from "@/api/openapi-schema";
import { ComposeForm } from "@/screens/compose/components/ComposeForm/ComposeForm";
import { OpenChangeEvent } from "@/utils/useDisclosure";

import { ComposerOverlay } from "./ComposerOverlay";

export type ThreadCreateModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onOpenChange: (event: OpenChangeEvent) => void;
  channelID: string;
  initialChannelList?: ChannelListOKResponse;
};

export function ThreadCreateModal({
  isOpen,
  onClose,
  onOpenChange,
  channelID,
  initialChannelList,
}: ThreadCreateModalProps) {
  return (
    <ComposerOverlay
      isOpen={isOpen}
      onClose={onClose}
      onOpenChange={onOpenChange}
      title="Create Post"
    >
      <ComposeForm
        channelID={channelID}
        onSuccess={onClose}
        skipDraftNavigation={true}
      />
    </ComposerOverlay>
  );
}
