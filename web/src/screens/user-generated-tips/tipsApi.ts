export interface SubmitTipArgs {
  caseId: string;
  topicId: string;
  topicTitle: string;
  text: string;
  hasImage?: boolean;
}

// TODO(backend): wire to the real user-tips endpoint when it lands.
export async function submitTip(
  args: SubmitTipArgs,
): Promise<{ hasError: boolean }> {
  try {
    console.debug("submitTip", args);
    return { hasError: false };
  } catch (error) {
    console.error("submitTip failed", error);
    return { hasError: true };
  }
}
