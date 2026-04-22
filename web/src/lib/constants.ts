/**
 * App-wide constants
 * Centralized file for static values, CDN URLs, and configuration constants.
 */

// Streak milestone images (Shopify CDN)
export const STREAK_IMAGES: Record<number, string> = {
  3: "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/3_days_streak.png?v=1775818613",
  7: "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/7_days_streak.png?v=1775818613",
  21: "https://cdn.shopify.com/s/files/1/0100/1622/7394/files/21_days_streak.png?v=1775818613",
};

export const DEFAULT_STREAK_IMAGE_KEY = 3;
