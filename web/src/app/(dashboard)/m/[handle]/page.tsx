import { UnreadyBanner } from "src/components/site/Unready";
import { ProfileScreen } from "src/screens/profile/ProfileScreen";

import { profileGet } from "@/api/openapi-server/profiles";
import { getServerSession } from "@/auth/server-session";
import { TempHandlePrompt } from "@/components/profile/TempHandlePrompt";

type Props = {
  params: Promise<{ handle: string }>;
};

export default async function Page(props: Props) {
  const params = await props.params;
  const { handle } = params;

  if (!handle || handle.startsWith("temp_")) {
    return <TempHandlePrompt />;
  }

  try {
    const session = await getServerSession();
    const { data } = await profileGet(handle);
    return <ProfileScreen initialSession={session} profile={data} />;
  } catch (e) {
    return <UnreadyBanner error="Something went wrong. Please try again later." />;
  }
}
