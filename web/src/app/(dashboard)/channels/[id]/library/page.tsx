import { channelGet } from "@/api/openapi-server/channels";
import { nodeList } from "@/api/openapi-server/nodes";
import { UnreadyBanner } from "@/components/site/Unready";
import { LibraryIndexScreen } from "@/screens/library/LibraryIndexScreen";

type Props = {
  params: Promise<{
    id: string;
  }>;
};

export default async function Page(props: Props) {
  try {
    const params = await props.params;
    const { data: channel } = await channelGet(params.id);
    const nodes = await nodeList({ channel: params.id });

    return (
      <LibraryIndexScreen
        nodes={nodes.data}
        channelID={params.id}
        channelName={channel.name}
      />
    );
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
