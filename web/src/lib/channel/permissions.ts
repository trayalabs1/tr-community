import { useChannelMemberList } from "@/api/openapi-client/channels";
import { useSession } from "@/auth";

/**
 * Hook to check if the current user has specific permissions in a channel
 *
 * Note: This hook uses SWR and must be used in client components only.
 * For server-side permission checks, see @/lib/channel/server-permissions
 */
export function useChannelPermissions(channelID: string) {
  const session = useSession();
  const { data: members } = useChannelMemberList(channelID);

  if (!session || !members) {
    return {
      isOwner: false,
      isAdmin: false,
      isModerator: false,
      isMember: false,
      canManageChannel: false,
      canManageMembers: false,
      canModerateContent: false,
      role: null,
    };
  }

  // Find current user's membership
  const membership = members.members.find(
    (m) => m.account.id === session.id
  );

  if (!membership) {
    return {
      isOwner: false,
      isAdmin: false,
      isModerator: false,
      isMember: false,
      canManageChannel: false,
      canManageMembers: false,
      canModerateContent: false,
      role: null,
    };
  }

  const role = membership.role;
  const isOwner = role === "owner";
  const isAdmin = role === "admin";
  const isModerator = role === "moderator";
  const isMember = role === "member";

  return {
    isOwner,
    isAdmin,
    isModerator,
    isMember,
    // Permissions
    canManageChannel: isOwner || isAdmin, // Can update channel settings
    canManageMembers: isOwner || isAdmin, // Can add/remove members
    canModerateContent: isOwner || isAdmin || isModerator, // Can moderate posts
    role,
  };
}
