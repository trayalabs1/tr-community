"use client";

import { Portal } from "@ark-ui/react";
import { useState } from "react";

import { MoreAction } from "src/components/site/Action/More";

import { DatagraphItemKind } from "@/api/openapi-schema";
import { CategoryMoveMenu } from "@/components/category/CategoryMoveMenu/CategoryMoveMenu";
import {
  ReportPostMenuItem,
  truncateBody,
} from "@/components/report/ReportPostMenuItem";
import { ThreadShareMenu } from "@/components/thread/ThreadShareMenu/ThreadShareMenu";
import { DeleteConfirmSheet } from "./DeleteConfirmSheet";
import { DeleteIcon } from "@/components/ui/icons/Delete";
import { EditIcon } from "@/components/ui/icons/Edit";
import { PinIcon, PinOffIcon } from "@/components/ui/icons/Pin";
import * as Menu from "@/components/ui/menu";
import { HStack } from "@/styled-system/jsx";

import { Props, useThreadMenu } from "./useThreadMenu";

export function ThreadMenu(props: Props) {
  const {
    isEditingEnabled,
    isMovingEnabled,
    isDeletingEnabled,
    canPinThread,
    isThreadPinned,
    canShareThread,
    handlers,
  } = useThreadMenu(props);

  const { thread, onRequestDelete } = props;

  const [showDeleteSheet, setShowDeleteSheet] = useState(false);

  const handleDeleteClick = onRequestDelete ?? (() => setShowDeleteSheet(true));

  return (
    <>
    <Menu.Root
      positioning={{
        shift: 32,
      }}
      lazyMount
    >
      <Menu.Trigger asChild>
        <MoreAction variant="subtle" size="xs" />
      </Menu.Trigger>

      <Portal>
        <Menu.Positioner>
          <Menu.Content minW="36" className="threadmenu__content">
            <Menu.ItemGroup id="group">
              <ReportPostMenuItem
                menuLabel="Report thread"
                targetKind={DatagraphItemKind.thread}
                targetId={thread.id}
                author={thread.author}
                headline={thread.title || "Untitled thread"}
                body={truncateBody(thread.description)}
              />

              {canPinThread && !isThreadPinned && (
                <Menu.Item value="pin" onClick={handlers.handlePinThread}>
                  <HStack gap="1">
                    <PinIcon /> Pin thread
                  </HStack>
                </Menu.Item>
              )}

              {canPinThread && isThreadPinned && (
                <Menu.Item value="unpin" onClick={handlers.handleUnpinThread}>
                  <HStack gap="1">
                    <PinOffIcon /> Unpin thread
                  </HStack>
                </Menu.Item>
              )}

              {isEditingEnabled && (
                <Menu.Item value="edit" onClick={handlers.handleEdit}>
                  <HStack gap="1">
                    <EditIcon /> Edit
                  </HStack>
                </Menu.Item>
              )}

              {isMovingEnabled && <CategoryMoveMenu thread={thread} />}

              {canShareThread && <ThreadShareMenu thread={thread} />}

              {isDeletingEnabled && (
                <Menu.Item
                  value="delete"
                  onClick={handleDeleteClick}
                  color="fg.destructive"
                  css={{ "& :where(svg)": { color: "fg.destructive" } }}
                >
                  <HStack gap="1">
                    <DeleteIcon /> Delete post
                  </HStack>
                </Menu.Item>
              )}
            </Menu.ItemGroup>
          </Menu.Content>
        </Menu.Positioner>
      </Portal>
    </Menu.Root>

    <style jsx global>{`
      /* Separate each action with a hairline, matching the design. */
      .threadmenu__content > [role="group"] > *:not(:last-child) {
        border-bottom: 1px solid var(--colors-border-default);
      }
    `}</style>

    {!onRequestDelete && (
      <DeleteConfirmSheet
        isOpen={showDeleteSheet}
        onOpenChange={setShowDeleteSheet}
        onConfirm={handlers.handleDelete}
      />
    )}
    </>
  );
}
