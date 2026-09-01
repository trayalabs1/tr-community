type ChannelAccess = {
  id: string;
};

// ChannelList only returns channels the caller can access, so presence in that
// list is the membership check.
//
// A thread is reachable via its home channel or via any channel it has been
// shared into, so a viewer who only shares a destination channel with it still
// has access. sharedToChannelIDs comes from the thread's shared_to_channel_ids.
// Returns the specific channel ID that grants access (home channel preferred
// over shares) so callers that need to route the viewer somewhere — not just
// answer yes/no — have a real destination.
export function resolveAccessibleChannelID(
  channelID: string | undefined,
  accessibleChannels: ChannelAccess[],
  sharedToChannelIDs?: string[],
): string | null {
  const isAccessible = (id: string | undefined): id is string =>
    !!id && accessibleChannels.some((c) => c.id === id);

  if (isAccessible(channelID)) return channelID;

  return (sharedToChannelIDs ?? []).find(isAccessible) ?? null;
}

export function hasChannelAccess(
  channelID: string | undefined,
  accessibleChannels: ChannelAccess[],
  sharedToChannelIDs?: string[],
): boolean {
  return (
    resolveAccessibleChannelID(
      channelID,
      accessibleChannels,
      sharedToChannelIDs,
    ) !== null
  );
}
