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

const COMMUNITY_CHANNEL_SLUGS = Object.values(COMMUNITY_CHANNEL_SLUG);

const TIPS_TARGET_META_KEY = "is_tips_target";

export async function submitTip({
  gender,
  topicId,
  topicTitle,
  text,
}: SubmitTipArgs): Promise<{ hasError: boolean }> {
  try {
    const { channels } = await channelList();
    const tagged = channels.find(
      (c) => c.meta?.[TIPS_TARGET_META_KEY] === true,
    );
    const memberships = channels.filter((c) =>
      COMMUNITY_CHANNEL_SLUGS.includes(c.slug),
    );
    const channel =
      tagged ??
      (memberships.length === 1
        ? memberships[0]
        : channels.find((c) => c.slug === COMMUNITY_CHANNEL_SLUG[gender]));
    if (!channel) {
      console.error("submitTip: no community channel for user", gender);
      return { hasError: true };
    }

    await channelThreadCreate(channel.id, {
      title: topicTitle,
      body: text.trim(),
      visibility: Visibility.review,
      meta: {
        post_category: "tip",
        type: topicId,
      },
    });
    return { hasError: false };
  } catch (error) {
    console.error("submitTip failed", error);
    return { hasError: true };
  }
}
