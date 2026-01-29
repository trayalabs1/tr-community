import { useCallback } from "react";
import type { MoEngageUser } from "./types";
import { moengage } from "./utils";

export function useMoengage() {
  const trackEvent = useCallback(
    (eventName: string, eventAttributes?: Record<string, unknown>) => {
      moengage.trackEvent(eventName, eventAttributes);
    },
    []
  );

  const setUserAttributes = useCallback((attributes: MoEngageUser) => {
    moengage.setUserAttributes(attributes);
  }, []);

  const logout = useCallback(() => {
    moengage.logout();
  }, []);

  const requestPushNotification = useCallback(() => {
    moengage.requestPushNotification();
  }, []);

  return {
    trackEvent,
    setUserAttributes,
    logout,
    requestPushNotification,
  };
}
