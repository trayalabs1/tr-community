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

const escapeHtml = (s: string): string =>
  s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

function buildTipBody(topicTitle: string, text: string): string {
  const body = escapeHtml(text.trim()).replace(/\n/g, "<br/>");
  return `<p><strong>${escapeHtml(topicTitle)}</strong></p><p>${body}</p>`;
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
      title: "Tip",
      body: buildTipBody(topicTitle, text),
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
