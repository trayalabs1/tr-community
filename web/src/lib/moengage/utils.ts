import type { MoEngageUser } from "./types";

export const moengage = {
  trackEvent: (eventName: string, eventAttributes?: Record<string, unknown>) => {
    if (typeof window !== "undefined" && window.Moengage) {
      try {
        window.Moengage.track_event?.(eventName, eventAttributes || {});
      } catch (error) {
        console.error(`[MoEngage] Error tracking event ${eventName}:`, error);
      }
    }
  },

  setUserAttributes: (attributes: MoEngageUser) => {
    if (typeof window !== "undefined" && window.Moengage) {
      try {
        if (attributes.email) {
          window.Moengage.add_email?.(attributes.email);
        }
        if (attributes.name) {
          window.Moengage.add_first_name?.(attributes.name);
        }
        if (attributes.mobile) {
          window.Moengage.add_mobile?.(attributes.mobile);
        }
      } catch (error) {
        console.error("[MoEngage] Error setting user attributes:", error);
      }
    }
  },

  logout: () => {
    if (typeof window !== "undefined" && window.Moengage) {
      try {
        window.Moengage.destroy_session?.();
      } catch (error) {
        console.error("[MoEngage] Error logging out:", error);
      }
    }
  },

  requestPushNotification: () => {
    if (typeof window !== "undefined" && window.Moengage) {
      try {
        window.Moengage.call_web_push?.();
      } catch (error) {
        console.error("[MoEngage] Error requesting push notification:", error);
      }
    }
  },

  checkStatus: () => {
    if (window.Moengage) {
      console.log("✅ MoEngage SDK is ready");
    } else {
      console.log("❌ MoEngage SDK not available");
    }
  },
};
