import { useChannelMembershipGet } from "@/api/openapi-client/channels";
import { useSession } from "@/auth";

export function useChannelPermissions(channelID: string) {
  const session = useSession();
  const { data: membershipData } = useChannelMembershipGet(channelID);

  if (!session || !membershipData) {
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

  const role = membershipData.role;
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
