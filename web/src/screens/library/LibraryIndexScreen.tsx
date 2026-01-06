"use client";

import { Unready } from "src/components/site/Unready";

import { Breadcrumbs } from "@/components/library/Breadcrumbs";
import { LibraryEmptyState } from "@/components/library/LibraryEmptyState";
import { NodeCardGrid } from "@/components/library/NodeCardList";
import { VStack } from "@/styled-system/jsx";

import { Props, useLibraryIndexScreen } from "./useLibraryIndexScreen";

export type LibraryIndexScreenProps = Props & {
  channelName?: string;
};

export function LibraryIndexScreen(props: LibraryIndexScreenProps) {
  const { ready, data, error } = useLibraryIndexScreen(props);

  if (!ready) return <Unready error={error} />;

  const { nodes } = data;

  return (
    <VStack gap="4">
      <Breadcrumbs
        libraryPath={[]}
        visibility="draft"
        create="show"
        channelName={props.channelName}
        channelID={props.channelID}
      />

      {nodes.data.nodes.length === 0 ? (
        <LibraryEmptyState />
      ) : (
        <NodeCardGrid
          libraryPath={[]}
          context="library"
          channelID={props.channelID}
          {...nodes.data}
        />
      )}
    </VStack>
  );
}
