"use client";

import { ModalDrawer } from "@/components/site/Modaldrawer/Modaldrawer";

import {
  ChannelCreateProps,
  ChannelCreateScreen,
} from "./ChannelCreateScreen";

export function ChannelCreateModal(props: ChannelCreateProps) {
  return (
    <>
      <ModalDrawer
        isOpen={props.isOpen}
        onClose={props.onClose}
        onOpenChange={props.onOpenChange}
        title="Create channel"
      >
        <ChannelCreateScreen {...props} />
      </ModalDrawer>
    </>
  );
}
