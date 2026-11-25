import { Account } from "@/api/openapi-schema";
import { hasPermission } from "@/utils/permissions";

/**
 * Check if user can create channels (requires ADMINISTRATOR permission)
 * Only administrators can create new channels
 *
 * Note: This is a server-safe function that doesn't use React hooks
 */
export function canCreateChannels(session?: Account): boolean {
  if (!session) return false;
  return hasPermission(session, "ADMINISTRATOR");
}