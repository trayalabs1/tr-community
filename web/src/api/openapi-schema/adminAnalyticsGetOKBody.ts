import type { AdminReplyTimeStat } from "./adminReplyTimeStat";
import type { ChannelAnalyticsStat } from "./channelAnalyticsStat";

export type AdminAnalyticsGetOKBody = {
  channel_onboardings: ChannelAnalyticsStat[];
  channel_posts: ChannelAnalyticsStat[];
  admin_reply_times: AdminReplyTimeStat[];
};
