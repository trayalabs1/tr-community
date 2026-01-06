"use client";

import { useNodeList } from "@/api/openapi-client/nodes";
import { NodeListResult, Visibility } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { hasPermission } from "@/utils/permissions";

export type LibraryNavigationTreeProps = {
  initialNodeList?: NodeListResult;
  currentNode: string | undefined;
  visibility: Visibility[];
  channelId?: string;
};

export function useLibraryNavigationTree({
  visibility,
  initialNodeList,
  channelId,
}: LibraryNavigationTreeProps) {
  const session = useSession();
  const { data, error } = useNodeList(
    {
      visibility,
      channel: channelId,
    },
    {
      swr: {
        fallbackData: initialNodeList,
      },
    },
  );
  if (!data) {
    return {
      ready: false as const,
      error,
    };
  }

  const canManageLibrary = hasPermission(session, "MANAGE_LIBRARY");

  return {
    ready: true as const,
    data,
    canManageLibrary,
  };
}
