import { ModalDrawer } from "src/components/site/Modaldrawer/Modaldrawer";

import { ChannelListOKResponse } from "@/api/openapi-schema";
import { ComposeForm } from "@/screens/compose/components/ComposeForm/ComposeForm";
import { OpenChangeEvent } from "@/utils/useDisclosure";

export type ThreadCreateModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onOpenChange: (event: OpenChangeEvent) => void;
  channelID: string;
  initialChannelList?: ChannelListOKResponse;
  streakCount?: number;
  rewardCoins?: number;
};

export function ThreadCreateModal({
  isOpen,
  onClose,
  onOpenChange,
  channelID,
  initialChannelList,
  streakCount,
  rewardCoins,
}: ThreadCreateModalProps) {
  return (
    <ModalDrawer
      isOpen={isOpen}
      onClose={onClose}
      onOpenChange={onOpenChange}
      title="Create Post"
    >
      <ComposeForm
        channelID={channelID}
        onSuccess={onClose}
        skipDraftNavigation={true}
        streakCount={streakCount}
        rewardCoins={rewardCoins}
      />
    </ModalDrawer>
  );
}
