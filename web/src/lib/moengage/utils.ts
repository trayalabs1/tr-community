import type { MoEngageUser } from "./types";

export const moengage = {
  trackEvent: (eventName: string, eventAttributes?: Record<string, unknown>) => {
    if (typeof window !== "undefined" && window.Moengage) {
      window.Moengage.push(["track_event", eventName, eventAttributes || {}]);
    }
  },

  setUserAttributes: (attributes: MoEngageUser) => {
    if (typeof window !== "undefined" && window.Moengage) {
      window.Moengage.push(["add_user_attribute", attributes]);
    }
  },

  logout: () => {
    if (typeof window !== "undefined" && window.Moengage) {
      window.Moengage.push(["logout"]);
    }
  },

  requestPushNotification: () => {
    if (typeof window !== "undefined" && window.Moengage) {
      window.Moengage.push(["moe_request_push"]);
    }
  },
};
