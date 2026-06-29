import { threadCreate } from "@/api/openapi-client/threads";
import { Visibility } from "@/api/openapi-schema";

export interface SubmitTipArgs {
  caseId: string;
  topicId: string;
  topicTitle: string;
  text: string;
  hasImage?: boolean;
}

// A tip is a community post tagged via `meta` so it can be filtered/surfaced as
// a tip — mirrors the BAH/feedback post creation in SharePostScreen. Created at
// `review` visibility so it lands in the moderation queue before going live.
export async function submitTip({
  topicId,
  topicTitle,
  text,
}: SubmitTipArgs): Promise<{ hasError: boolean }> {
  try {
    await threadCreate({
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
