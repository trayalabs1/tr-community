import { UnreadyBanner } from "src/components/site/Unready";

import { channelGet, channelThreadGet } from "@/api/openapi-server/channels";
import { getServerSession } from "@/auth/server-session";

import { ChannelThreadScreen } from "./ChannelThreadScreen";

type Props = {
  params: Promise<{
    id: string;
    thread_slug: string;
  }>;
  searchParams: Promise<{
    page?: string;
  }>;
};

export default async function Page(props: Props) {
  try {
    const params = await props.params;
    const searchParams = await props.searchParams;
    const session = await getServerSession();

    const [channelResult, threadResult] = await Promise.all([
      channelGet(params.id),
      channelThreadGet(params.id, params.thread_slug, {
        page: searchParams.page,
      }),
    ]);

    return (
      <ChannelThreadScreen
        session={session}
        channel={channelResult.data}
        channelId={params.id}
        slug={params.thread_slug}
        thread={threadResult.data}
        initialPage={searchParams.page ? parseInt(searchParams.page) : 1}
      />
    );
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
