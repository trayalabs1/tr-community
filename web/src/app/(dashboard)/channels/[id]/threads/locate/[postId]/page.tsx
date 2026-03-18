import { redirect } from "next/navigation";

import { PostLocationKind } from "@/api/openapi-schema";
import { postLocationGet } from "@/api/openapi-server/posts";
import { WEB_ADDRESS } from "@/config";

export type Props = {
  params: Promise<{
    id: string;
    postId: string;
  }>;
  searchParams: Promise<{
    [key: string]: string | string[] | undefined;
  }>;
};

export default async function ChannelThreadLocatePage(props: Props) {
  const { id: channelId, postId } = await props.params;
  const searchParams = await props.searchParams;

  const { data } = await postLocationGet({ id: postId });

  const url = new URL(
    `/channels/${channelId}/threads/${data.slug}`,
    WEB_ADDRESS,
  );

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

  url.hash = postId;

  redirect(url.toString());
}
