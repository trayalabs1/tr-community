"use client";

import { useCallback } from "react";
import { useAccountGet } from "src/api/openapi-client/accounts";
import { moengage } from "./utils";
import type { EventName } from "./events";

interface ProfileMetadata extends Record<string, unknown> {
  case_id?: string;
}

interface AccountWithProfile extends Record<string, unknown> {
  id?: string;
  meta?: ProfileMetadata;
}

export function useEventTracking() {
  const { data: account } = useAccountGet();

  const getCaseId = useCallback(() => {
    if (!account) {
      return undefined;
    }

    const accountData = account as unknown as AccountWithProfile;
    return accountData.meta?.case_id || accountData.id || undefined;
  }, [account]);

  const trackEvent = useCallback(
    (eventName: EventName, attributes?: Record<string, unknown>) => {
      const caseId = getCaseId();
      setTimeout(() => {
        moengage.trackEvent(eventName, {
          case_id: caseId,
          ...attributes,
        });
      }, 100);
    },
    [getCaseId]
  );

  // Convenience methods for common patterns (optional, for backward compatibility)
  const trackOnboardingLanded = useCallback(() => {
    trackEvent("community_onboarding_landed");
  }, [trackEvent]);

  const trackEnterClicked = useCallback(() => {
    trackEvent("community_enter_clicked");
  }, [trackEvent]);

  const trackNameFilled = useCallback((communityName?: string) => {
    trackEvent("community_name_filled", {
      community_name: communityName,
    });
  }, [trackEvent]);

  const trackLandedHome = useCallback(() => {
    trackEvent("community_landed_home");
  }, [trackEvent]);

  const trackProfileClicked = useCallback(() => {
    trackEvent("community_home_profile_clicked");
  }, [trackEvent]);

  const trackSavedClicked = useCallback(() => {
    trackEvent("community_home_saved_clicked");
  }, [trackEvent]);

  const trackNotificationsClicked = useCallback(() => {
    trackEvent("community_home_notifications_clicked");
  }, [trackEvent]);

  const trackNameChanged = useCallback((newName?: string) => {
    trackEvent("community_name_changed", {
      new_community_name: newName,
    });
  }, [trackEvent]);

  const trackSearchClicked = useCallback(() => {
    trackEvent("community_search_clicked");
  }, [trackEvent]);

  const trackSearchDone = useCallback(
    (query?: string, resultsCount?: number) => {
      trackEvent("community_search_done", {
        search_query: query,
        results_count: resultsCount,
      });
    },
    [trackEvent]
  );

  const trackInfoClicked = useCallback(() => {
    trackEvent("community_info_clicked");
  }, [trackEvent]);

  const trackPostClicked = useCallback(
    (channelId?: string, postType?: string) => {
      trackEvent("community_post_clicked", {
        channel_id: channelId,
        post_type: postType,
      });
    },
    [trackEvent]
  );

  const trackSubmitForReview = useCallback(
    (postLength?: number, hasAttachments?: boolean, hasMentions?: boolean, channelId?: string) => {
      trackEvent("community_submit_for_review", {
        post_length: postLength,
        has_attachments: hasAttachments,
        has_mentions: hasMentions,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  const trackCardReply = useCallback(
    (postId?: string, replyLength?: number, channelId?: string) => {
      trackEvent("community_card_reply", {
        post_id: postId,
        reply_length: replyLength,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  const trackCardRepliedTo = useCallback(
    (postId?: string, replierId?: string, channelId?: string) => {
      trackEvent("community_card_replied_to", {
        post_id: postId,
        replier_id: replierId,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  const trackCardLike = useCallback((postId?: string, channelId?: string) => {
    trackEvent("community_card_like", {
      post_id: postId,
      channel_id: channelId,
    });
  }, [trackEvent]);

  const trackCardLikedTo = useCallback(
    (postId?: string, likerId?: string, channelId?: string) => {
      trackEvent("community_card_liked_to", {
        post_id: postId,
        liker_id: likerId,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  const trackCardSave = useCallback(
    (postId?: string, action?: "save" | "unsave", channelId?: string) => {
      trackEvent("community_card_save", {
        post_id: postId,
        action,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  const trackAdminApproved = useCallback(
    (postId?: string, approvedPostOwnerId?: string, channelId?: string) => {
      trackEvent("community_admin_approved", {
        post_id: postId,
        approved_post_owner_id: approvedPostOwnerId,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  const trackAdminReplied = useCallback(
    (postId?: string, replyLength?: number, postOwnerId?: string, channelId?: string) => {
      trackEvent("community_admin_replied", {
        post_id: postId,
        reply_length: replyLength,
        post_owner_id: postOwnerId,
        channel_id: channelId,
      });
    },
    [trackEvent]
  );

  return {
    trackEvent,
    trackOnboardingLanded,
    trackEnterClicked,
    trackNameFilled,
    trackLandedHome,
    trackProfileClicked,
    trackSavedClicked,
    trackNotificationsClicked,
    trackNameChanged,
    trackSearchClicked,
    trackSearchDone,
    trackInfoClicked,
    trackPostClicked,
    trackSubmitForReview,
    trackCardReply,
    trackCardRepliedTo,
    trackCardLike,
    trackCardLikedTo,
    trackCardSave,
    trackAdminApproved,
    trackAdminReplied,
  };
}
