"use client";

import { useEffect } from "react";

export function MoEngageInitializer() {
  useEffect(() => {
    if (typeof window !== "undefined") {
      // Register service worker for MoEngage push notifications
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker
          .register("/serviceworker.js")
          .then(() => {
            console.log("[MoEngage] Service Worker registered successfully");
          })
          .catch((error) => {
            console.error("[MoEngage] Service Worker registration failed:", error);
          });
      }
    }
  }, []);

  return null;
}
