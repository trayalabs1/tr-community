import { UnreadyBanner } from "src/components/site/Unready";

import { channelGet } from "@/api/openapi-server/channels";
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
    const session = await getServerSession();
    const { data: channel } = await channelGet(params.id);
    const { data: notifications } = await notificationList({ status: ["unread"], page: "1" });
    const { data: collections } = await collectionList({});

    const hasUnreadNotifications = (notifications?.notifications?.length ?? 0) > 0;
    const bookmarkCount = collections?.collections?.length ?? 0;

    return (
      <ChannelScreen
        session={session}
        channel={channel}
        hasUnreadNotifications={hasUnreadNotifications}
        bookmarkCount={bookmarkCount}
      />
    );
  } catch (e) {
    return <UnreadyBanner error={e} />;
  }
}
