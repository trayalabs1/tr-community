import { z } from "zod";

import { channelGet } from "@/api/openapi-server/channels";
import { threadGet } from "@/api/openapi-server/threads";
import { getServerSession } from "@/auth/server-session";
import { getSettings } from "@/lib/settings/settings-server";
import { ThreadScreen } from "@/screens/thread/ThreadScreen/ThreadScreen";

export type Props = {
  params: Promise<{
    id: string;
    slug: string;
  }>;
  searchParams: Promise<Query>;
};

const QuerySchema = z.object({
  page: z
    .string()
    .transform((v) => parseInt(v, 10))
    .optional(),
});

type Query = z.infer<typeof QuerySchema>;

export default async function Page(props: Props) {
  const { id: channelID, slug } = await props.params;
  const searchParams = await props.searchParams;

  const { page } = QuerySchema.parse(searchParams);

  const { data: thread } = await threadGet(slug, {
    page: page?.toString(),
  });

  const { data: channel } = await channelGet(channelID);

  const session = await getServerSession();

  return (
    <ThreadScreen
      initialSession={session}
      initialPage={page}
      slug={slug}
      thread={thread}
      channelID={channelID}
      channelName={channel.name}
    />
  );
}

export async function generateMetadata(props: Props) {
  const params = await props.params;
  try {
    const settings = await getSettings();
    const { data: thread } = await threadGet(params.slug);
    const { data: channel } = await channelGet(params.id);

    return {
      title: `${thread.title} | ${channel.name} | ${settings.title}`,
      description: thread.description,
    };
  } catch (e) {
    return {
      title: "Thread Not Found",
      description: "The thread you are looking for does not exist.",
    };
  }
}
