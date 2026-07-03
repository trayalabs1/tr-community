import type { Channel } from "@/api/openapi-schema";
import { PINNED_CHANNEL_SLUGS } from "./pinned";

const TIPS_TARGET_META_KEY = "is_tips_target";
const FEMALE_CHANNEL_RE =
  /heroines|divas|icons|anchors|elites|champions|queens|women/i;

// The community channel a user's tip posts into: an explicitly tagged channel,
// else their cohort ("month …") channel, else a pinned/first channel.
export function pickTipsChannel(channels: Channel[]): Channel | undefined {
  return (
    channels.find((c) => c.meta?.[TIPS_TARGET_META_KEY] === true) ??
    channels.find((c) => c.name?.toLowerCase().includes("month")) ??
    channels.find((c) => PINNED_CHANNEL_SLUGS.includes(c.slug)) ??
    channels[0]
  );
}

// Gender inferred from the channel's cohort naming — used only for theming.
export function inferChannelGender(channel: Channel | undefined): "M" | "F" {
  if (
    channel &&
    (FEMALE_CHANNEL_RE.test(channel.slug) ||
      FEMALE_CHANNEL_RE.test(channel.name ?? ""))
  ) {
    return "F";
  }
  return "M";
}
