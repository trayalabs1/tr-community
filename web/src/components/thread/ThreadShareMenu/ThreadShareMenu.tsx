import { Portal } from "@ark-ui/react";

import { useChannelList } from "@/api/openapi-client/channels";
import { ThreadReference } from "@/api/openapi-schema";
import { Unready } from "@/components/site/Unready";
import { Input } from "@/components/ui/input";
import { CategoryIcon } from "@/components/ui/icons/Category";
import { CheckIcon } from "@/components/ui/icons/Check";
import { SubmenuIcon } from "@/components/ui/icons/Submenu";
import { ShareIcon } from "@/components/ui/icons/Share";
import * as Menu from "@/components/ui/menu";
import { Button } from "@/components/ui/button";
import { HStack, VStack } from "@/styled-system/jsx";

import { TRAYA_COLORS } from "@/theme/traya-colors";

import { useThreadShareMenu } from "./useThreadShareMenu";

type Props = {
  thread: ThreadReference;
};

export function ThreadShareMenu({ thread }: Props) {
  const { selected, subtitle, handlers } = useThreadShareMenu({ thread });

  return (
    <Menu.Root
      size="xs"
      positioning={{ placement: "right-start", gutter: -2 }}
      closeOnSelect={false}
    >
      <Menu.TriggerItem justifyContent="space-between">
        <HStack gap="1">
          <ShareIcon />
          Share to another channel
        </HStack>
        <SubmenuIcon />
      </Menu.TriggerItem>

      <Portal>
        <Menu.Positioner>
          <LazyLoadedThreadShareMenuContent
            selected={selected}
            subtitle={subtitle}
            onToggleChannel={handlers.handleToggleChannel}
            onSubtitleChange={handlers.handleSubtitleChange}
            onConfirmShare={handlers.handleConfirmShare}
          />
        </Menu.Positioner>
      </Portal>
    </Menu.Root>
  );
}

type ContentProps = {
  selected: Set<string>;
  subtitle: string;
  onToggleChannel: (channelID: string) => void;
  onSubtitleChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onConfirmShare: () => Promise<void>;
};

function LazyLoadedThreadShareMenuContent({
  selected,
  subtitle,
  onToggleChannel,
  onSubtitleChange,
  onConfirmShare,
}: ContentProps) {
  const { data, error } = useChannelList();
  if (!data) {
    return <Unready error={error} />;
  }

  const { channels } = data;

  if (channels.length === 0) {
    return (
      <Menu.Content minW="64" userSelect="none">
        <Menu.ItemGroup id="share-no-channels">
          <Menu.ItemGroupLabel>No other channels to share to</Menu.ItemGroupLabel>
        </Menu.ItemGroup>
      </Menu.Content>
    );
  }

  return (
    <Menu.Content minW="64" userSelect="none">
      <Menu.ItemGroup id="share">
        <Menu.ItemGroupLabel>Share to channels</Menu.ItemGroupLabel>

        <Menu.Separator />

        {channels.map((c) => {
          const isSelected = selected.has(c.id);
          return (
            <Menu.CheckboxItem
              key={c.id}
              value={c.id}
              checked={isSelected}
              onCheckedChange={() => onToggleChannel(c.id)}
              closeOnSelect={false}
              justifyContent="space-between"
            >
              <HStack gap="1" minW="0">
                <CategoryIcon />
                {c.name}
              </HStack>
              {isSelected && (
                <CheckIcon
                  width="4"
                  style={{ color: TRAYA_COLORS.primary }}
                />
              )}
            </Menu.CheckboxItem>
          );
        })}

        <Menu.Separator />

        <VStack gap="2" alignItems="stretch" px="2" py="1.5">
          <Input
            size="xs"
            placeholder="Subtitle (optional)"
            aria-label="Subtitle for the shared thread"
            value={subtitle}
            onChange={onSubtitleChange}
            onKeyDown={(e) => e.stopPropagation()}
          />

          <ConfirmShareButton selected={selected} onConfirmShare={onConfirmShare} />
        </VStack>
      </Menu.ItemGroup>
    </Menu.Content>
  );
}

type ConfirmProps = {
  selected: Set<string>;
  onConfirmShare: () => Promise<void>;
};

function ConfirmShareButton({ selected, onConfirmShare }: ConfirmProps) {
  const count = selected.size;

  return (
    <Button
      size="xs"
      variant="subtle"
      disabled={count === 0}
      onClick={onConfirmShare}
    >
      {count === 0 ? "Select channels" : `Share to ${count} channel${count === 1 ? "" : "s"}`}
    </Button>
  );
}
