import { channelGet } from "@/api/openapi-server/channels";
import { ChannelScreenContextPane } from "@/screens/channel/ChannelScreenContextPane";

export default async function Page(props: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await props.params;

  try {
    const { data: channel } = await channelGet(id);

    return <ChannelScreenContextPane channel={channel} />;
  } catch (e) {
    console.error(e);
    return null;
  }
}
