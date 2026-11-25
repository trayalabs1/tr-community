import { UnreadyBanner } from "src/components/site/Unready";

import { channelGet } from "@/api/openapi-server/channels";
import { getServerSession } from "@/auth/server-session";

import { ChannelSettingsScreen } from "./ChannelSettingsScreen";

type Props = {
  params: Promise<{
    id: string;
  }>;
};

export default async function Page(props: Props) {
  try {
    const params = await props.params;
    const session = await getServerSession();
    const { data: channel } = await channelGet(params.id);

    return <ChannelSettingsScreen session={session} channel={channel} />;
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
