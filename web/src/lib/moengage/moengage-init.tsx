"use client";

import { useEffect } from "react";

export function MoEngageInitializer() {
  useEffect(() => {
    const appId = process.env["NEXT_PUBLIC_MOENGAGE_APP_ID"];

    if (!appId) {
      console.warn("MoEngage App ID is not configured");
      return;
    }

    if (typeof window !== "undefined" && window.Moengage) {
      window.Moengage.push([
        "init",
        {
          app_id: appId,
          swPath: "/serviceworker.js",
        },
      ]);
    }
  }, []);

  return null;
}
