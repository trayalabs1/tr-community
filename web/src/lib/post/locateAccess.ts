type ChannelAccess = {
  id: string;
};

// ChannelList only returns channels the caller can access, so presence in that
// list is the membership check.
//
// A thread is reachable via its home channel or via any channel it has been
// shared into, so a viewer who only shares a destination channel with it still
// has access. sharedToChannelIDs comes from the thread's shared_to_channel_ids.
export function hasChannelAccess(
  channelID: string | undefined,
  accessibleChannels: ChannelAccess[],
  sharedToChannelIDs?: string[],
): boolean {
  const isAccessible = (id: string | undefined) =>
    !!id && accessibleChannels.some((c) => c.id === id);

  if (isAccessible(channelID)) return true;

  return (sharedToChannelIDs ?? []).some(isAccessible);
}
