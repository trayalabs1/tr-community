"use server";

import { revalidatePath } from "next/cache";

export async function revalidateChannels(channelId?: string) {
  revalidatePath("/channels");
  if (channelId) {
    revalidatePath(`/channels/${channelId}`);
  }
}
