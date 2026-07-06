import { UnreadyBanner } from "src/components/site/Unready";

import {
  channelGet,
  channelCategoryList,
  channelThreadList,
} from "@/api/openapi-server/channels";
import { notificationList } from "@/api/openapi-server/notifications";
import { collectionList } from "@/api/openapi-server/collections";
import { getServerSession } from "@/auth/server-session";

import { ChannelScreen } from "./ChannelScreen";

type Props = {
  params: Promise<{
    id: string;
  }>;
};

export default async function Page(props: Props) {
  try {
    const params = await props.params;
    const [
      session,
      { data: channel },
      { data: notifications },
      { data: collections },
      { data: threads },
      { data: categories },
    ] = await Promise.all([
      getServerSession(),
      channelGet(params.id),
      notificationList({ status: ["unread"], page: "1" }),
      collectionList({}),
      channelThreadList(params.id, { page: "1" }),
      channelCategoryList(params.id),
    ]);

    const hasUnreadNotifications = (notifications?.notifications?.length ?? 0) > 0;
    const bookmarkCount = collections?.collections?.length ?? 0;

    return (
      <ChannelScreen
        session={session}
        channel={channel}
        hasUnreadNotifications={hasUnreadNotifications}
        bookmarkCount={bookmarkCount}
        initialThreads={threads}
        initialCategories={categories}
      />
    );
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
