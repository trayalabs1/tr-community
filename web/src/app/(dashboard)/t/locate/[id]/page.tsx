import { redirect } from "next/navigation";

import { PostLocationKind } from "@/api/openapi-schema";
import { channelList } from "@/api/openapi-server/channels";
import { postLocationGet } from "@/api/openapi-server/posts";
import { threadGet } from "@/api/openapi-server/threads";
import { WEB_ADDRESS } from "@/config";
import { hasChannelAccess } from "@/lib/post/locateAccess";

export type Props = {
  params: Promise<{
    id: string;
  }>;
  searchParams: Promise<{
    [key: string]: string | string[] | undefined;
  }>;
};

// Returns null when the post is missing, its thread is unreadable, or the viewer
// can reach neither its home channel nor any channel it's been shared into. Kept
// separate from the page body so redirect()'s control-flow throw isn't caught
// here.
async function locatePost(id: string) {
  try {
    const { data: location } = await postLocationGet({ id });
    if (!location?.slug) return null;

    const [{ data: thread }, { data: channels }] = await Promise.all([
      threadGet(location.slug),
      channelList(),
    ]);

    const accessible = channels?.channels ?? [];
    if (
      !hasChannelAccess(
        thread?.channel_id,
        accessible,
        thread?.shared_to_channel_ids,
      )
    ) {
      return null;
    }

    return location;
  } catch {
    return null;
  }
}

export default async function LocatePage(props: Props) {
  const { id } = await props.params;
  const searchParams = await props.searchParams;

  // Resolve the post and confirm the viewer can see its channel before sending
  // them onward — /t/[slug] has no guard of its own, so an unreachable post
  // would render an error screen. Deep links arrive from notifications and
  // marketing banners, where the feed is a better landing than an error.
  const data = await locatePost(id);
  if (!data) {
    redirect("/channels");
  }

  const url = new URL(`/t/${data.slug}`, WEB_ADDRESS);

  // we pass through any parameters from the original call to the final URL
  Object.entries(searchParams).forEach(([key, value]) => {
    if (value === undefined) return;
    if (typeof value === "string") {
      url.searchParams.set(key, value);
    } else if (Array.isArray(value)) {
      value.forEach((v) => url.searchParams.append(key, v));
    }
  });

  if (data.kind === PostLocationKind.thread) {
    redirect(url.toString());
  }

  if (data.page && data.page > 1) {
    url.searchParams.set("page", data.page.toString());
  }

  url.hash = id;

  redirect(url.toString());
}
