import type { MoEngageUser } from "./types";

type QueuedEvent = {
  eventName: string;
  eventAttributes?: Record<string, unknown>;
};

let isIdentified = false;
let eventQueue: QueuedEvent[] = [];

function waitForMoengage(callback: () => void, maxAttempts = 50) {
  let attempts = 0;
  const check = () => {
    if (typeof window !== "undefined" && window.Moengage) {
      callback();
    } else if (attempts < maxAttempts) {
      attempts++;
      setTimeout(check, 100);
    }
  };
  check();
}

function flushEventQueue() {
  if (!isIdentified) return;

  while (eventQueue.length > 0) {
    const event = eventQueue.shift();
    if (event && window.Moengage) {
      try {
        window.Moengage.track_event?.(event.eventName, event.eventAttributes || {});
      } catch (error) {
        console.error(`[MoEngage] Error tracking queued event ${event.eventName}:`, error);
      }
    }
  }
}

export const moengage = {
  identify: (uniqueId: string) => {
    if (!uniqueId) return;

    waitForMoengage(() => {
      try {
        window.Moengage?.add_unique_user_id?.(uniqueId);
        isIdentified = true;
        flushEventQueue();
      } catch (error) {
        console.error("[MoEngage] Error identifying user:", error);
      }
    });
  },

  trackEvent: (eventName: string, eventAttributes?: Record<string, unknown>) => {
    if (!isIdentified) {
      eventQueue.push({ eventName, eventAttributes });
      return;
    }

    waitForMoengage(() => {
      try {
        window.Moengage?.track_event?.(eventName, eventAttributes || {});
      } catch (error) {
        console.error(`[MoEngage] Error tracking event ${eventName}:`, error);
      }
    });
  },

  setUserAttributes: (attributes: MoEngageUser) => {
    waitForMoengage(() => {
      try {
        if (attributes.email) {
          window.Moengage?.add_email?.(attributes.email);
        }
        if (attributes.name) {
          window.Moengage?.add_first_name?.(attributes.name);
        }
        if (attributes.mobile) {
          window.Moengage?.add_mobile?.(attributes.mobile);
        }
      } catch (error) {
        console.error("[MoEngage] Error setting user attributes:", error);
      }
    });
  },

  logout: () => {
    waitForMoengage(() => {
      try {
        window.Moengage?.destroy_session?.();
        isIdentified = false;
        eventQueue = [];
      } catch (error) {
        console.error("[MoEngage] Error logging out:", error);
      }
    });
  },

  requestPushNotification: () => {
    waitForMoengage(() => {
      try {
        window.Moengage?.call_web_push?.();
      } catch (error) {
        console.error("[MoEngage] Error requesting push notification:", error);
      }
    });
  },

  isIdentified: () => isIdentified,

  getQueueLength: () => eventQueue.length,
};
