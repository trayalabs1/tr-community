export const EVENTS = {
  community_onboarding_landed: {},
  community_enter_clicked: {},
  community_name_filled: { optional: ["community_name"] },
  community_landed_home: {},
  community_home_profile_clicked: {},
  community_home_saved_clicked: {},
  community_home_notifications_clicked: {},
  community_name_changed: { optional: ["new_community_name"] },
  community_search_clicked: {},
  community_search_done: { optional: ["search_query", "results_count"] },
  community_info_clicked: {},
  community_post_clicked: { optional: ["channel_id", "post_type"] },
  community_submit_for_review: {
    optional: ["post_length", "has_attachments", "has_mentions"],
  },
  community_card_reply: { optional: ["post_id", "reply_length"] },
  community_card_replied_to: { optional: ["post_id", "replier_id"] },
  community_card_like: { optional: ["post_id"] },
  community_card_liked_to: { optional: ["post_id", "liker_id"] },
  community_card_save: { optional: ["post_id", "action"] },
  community_admin_approved: {
    optional: ["post_id", "approved_post_owner_id"],
  },
  community_admin_replied: {
    optional: ["post_id", "reply_length", "post_owner_id"],
  },
} as const;

export type EventName = keyof typeof EVENTS;
