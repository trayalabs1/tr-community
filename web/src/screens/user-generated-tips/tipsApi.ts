import { channelList, channelThreadCreate } from "@/api/openapi-client/channels";
import { Visibility } from "@/api/openapi-schema";
import type { Gender } from "./tipsData";

export interface SubmitTipArgs {
  caseId: string;
  gender: Gender;
  topicId: string;
  topicTitle: string;
  text: string;
  hasImage?: boolean;
}

const COMMUNITY_CHANNEL_SLUG: Record<Gender, string> = {
  F: "traya-womens-community",
  M: "traya-explorers",
};

export async function submitTip({
  gender,
  topicId,
  topicTitle,
  text,
}: SubmitTipArgs): Promise<{ hasError: boolean }> {
  const slug = COMMUNITY_CHANNEL_SLUG[gender];
  try {
    const { channels } = await channelList();
    const channel = channels.find((c) => c.slug === slug);
    if (!channel) {
      console.error("submitTip: community channel not found", slug);
      return { hasError: true };
    }

    await channelThreadCreate(channel.id, {
      title: topicTitle,
      body: text.trim(),
      visibility: Visibility.review,
      meta: {
        post_category: "tip",
        topic: topicId,
      },
    });
    return { hasError: false };
  } catch (error) {
    console.error("submitTip failed", error);
    return { hasError: true };
  }
}
