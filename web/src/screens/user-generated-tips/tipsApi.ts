import { channelList, channelThreadCreate } from "@/api/openapi-client/channels";
import { Visibility } from "@/api/openapi-schema";
import { pickTipsChannel } from "@/lib/channel/tipsChannel";

export interface SubmitTipArgs {
  caseId: string;
  topicId: string;
  topicTitle: string;
  text: string;
  hasImage?: boolean;
}

export async function submitTip({
  topicId,
  topicTitle,
  text,
}: SubmitTipArgs): Promise<{ hasError: boolean }> {
  try {
    const { channels } = await channelList();
    const channel = pickTipsChannel(channels);
    if (!channel) {
      console.error("submitTip: user has no community channel");
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
