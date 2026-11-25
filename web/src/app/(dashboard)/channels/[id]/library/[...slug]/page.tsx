import { Metadata } from "next";

import { NodeListResult, NodeWithChildren } from "@/api/openapi-schema";
import { channelGet } from "@/api/openapi-server/channels";
import { nodeGet, nodeListChildren } from "@/api/openapi-server/nodes";
import { getTargetSlug } from "@/components/library/utils";
import { WEB_ADDRESS } from "@/config";
import {
  LibraryPageBlockTypeDirectory,
  parseNodeMetadata,
} from "@/lib/library/metadata";
import { getSettings } from "@/lib/settings/settings-server";
import { LibraryPageScreen } from "@/screens/library/LibraryPageScreen/LibraryPageScreen";
import { Params, ParamsSchema } from "@/screens/library/library-path";

export type Props = {
  params: Promise<{
    id: string;
    slug: string[];
  }>;
};

export default async function Page(props: Props) {
  const params = await props.params;
  const { slug } = ParamsSchema.parse({ slug: params.slug });

  const targetSlug = getTargetSlug(slug);

  if (!targetSlug) {
    throw new Error("Library page not found");
  }

  const { data: channel } = await channelGet(params.id);
  const { data } = await nodeGet(targetSlug, undefined, {
    cache: "no-store",
    next: {
      tags: ["library", "node"],
      revalidate: 1,
    },
  });

  const children = await maybeGetChildren(data);

  return (
    <LibraryPageScreen
      node={data}
      childNodes={children}
      channelID={params.id}
      channelName={channel.name}
    />
  );
}

export async function generateMetadata(props: Props) {
  try {
    const params = await props.params;
    const { slug } = ParamsSchema.parse({ slug: params.slug });

    const targetSlug = getTargetSlug(slug);

    if (!targetSlug) {
      throw new Error("Library page not found");
    }

    const settings = await getSettings();

    const { data } = await nodeGet(targetSlug);

    return {
      title: `${data.name} | ${settings.title}`,
      description: data.description,
      openGraph: {
        images: [`${WEB_ADDRESS}/l/og?slug=${targetSlug}&t=${data.updatedAt}`],
      },
    } satisfies Metadata;
  } catch (e) {
    return {
      title: "Page not found",
      description: "The page you are looking for does not exist.",
    };
  }
}

async function maybeGetChildren(
  node: NodeWithChildren,
): Promise<NodeListResult | undefined> {
  if (!node.hide_child_tree) {
    return;
  }

  const directory = parseNodeMetadata(node.meta).layout?.blocks.find(
    (b): b is LibraryPageBlockTypeDirectory => b.type === "directory",
  );
  if (!directory) {
    return;
  }

  const { data } = await nodeListChildren(node.slug);
  return data;
}
