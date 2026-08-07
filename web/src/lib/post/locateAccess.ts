type ChannelAccess = {
  id: string;
};

// ChannelList only returns channels the caller can access, so presence in that
// list is the membership check.
export function hasChannelAccess(
  channelID: string | undefined,
  accessibleChannels: ChannelAccess[],
): boolean {
  if (!channelID) return false;

  return accessibleChannels.some((c) => c.id === channelID);
}
