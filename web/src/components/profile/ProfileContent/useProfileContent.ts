import { Account, PublicProfile } from "src/api/openapi-schema";

import { useThreadList } from "@/api/openapi-client/threads";

export type Props = {
  session?: Account;
  profile: PublicProfile;
};

export function useProfileContent({ session, profile }: Props) {
  const threads = useThreadList({ author: profile.handle });

  if (!threads.data) {
    return { ready: false as const, error: threads.error };
  }

  const isSelf = session?.id === profile.id;

  return {
    ready: true as const,
    isSelf,
    data: {
      threads: threads.data.threads,
    },
  };
}
