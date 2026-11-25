import { UnreadyBanner } from "src/components/site/Unready";

import {
  channelCategoryGet,
  channelGet,
} from "@/api/openapi-server/channels";
import { getServerSession } from "@/auth/server-session";

import { ChannelCategoryScreen } from "./ChannelCategoryScreen";

type Props = {
  params: Promise<{ id: string; slug: string }>;
};

export default async function Page(props: Props) {
  try {
    const params = await props.params;
    const session = await getServerSession();
    const { data: channel } = await channelGet(params.id);
    const { data: category } = await channelCategoryGet(params.id, params.slug);

    return (
      <ChannelCategoryScreen
        session={session}
        channel={channel}
        category={category}
      />
    );
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
