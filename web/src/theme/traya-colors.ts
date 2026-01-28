/**
 * Traya Brand Colors
 * Centralized color definitions for consistent styling across the app
 *
 * Color Palette:
 * - Primary: #329866 (Main brand color, used for text and key elements)
 * - Secondary: #B2E6CD (Used for backgrounds, icons, and hover states)
 * - Tertiary: #DCF4E8 (Light background, used for hover and highlights)
 * - Gradient: linear-gradient(180deg, #3EBC7F 17.69%, #329866 100%) (Primary to Secondary gradient)
 * - Heart: #F04343 (Used for like/heart actions)
 *
 * Font: Nunito Sans
 */

export const TRAYA_COLORS = {
  // Primary brand color - used for CTAs and main actions
  primary: "#329866",

  // Secondary color - used for user capsules and badges
  secondary: "#B2E6CD",

  // Tertiary color - light background and inactive states
  tertiary: "#DCF4E8",

  // Gradient - primary to secondary gradient
  gradient: "linear-gradient(180deg, #3EBC7F 17.69%, #329866 100%)",

  // Accent colors
  heart: "#F04343",

  // Color Usage by Screen/Area (Following exact design flow)
  screens: {
    home: {
      profileIcon: "gradient",      // Use gradient
      savedDot: "primary",           // #329866
    },
    saved: {
      profileIcon: "tertiary",       // #DCF4E8
      userPostCapsule: "secondary",  // #B2E6CD
      userReplyCapsule: "secondary", // #B2E6CD
    },
    profile: {
      profileIcon: "gradient",       // Use gradient
      memberAdminCapsule: "secondary", // #B2E6CD
      memberSinceLikes: "tertiary",  // #DCF4E8
    },
    threads: {
      productUpdatesCapsule: "secondary", // #B2E6CD
    },
    search: {
      profileIcon: "tertiary",       // #DCF4E8
      filterIcon: "tertiary",        // #DCF4E8
      capsulesList: "secondary",     // #B2E6CD
    },
    capsules: {
      insideCards: "secondary",      // #B2E6CD
      insideComments: "secondary",   // #B2E6CD
    },
    creation: {
      submitForReview: "primary",    // #329866
      createCollection: "primary",   // #329866
    },
  },

  // Neutral colors for complementary styling
  neutral: {
    light: "#F3F4F6",
    border: "#E5E7EB",
    text: "#6B7280",
    textMuted: "#9CA3AF",
    white: "#ffffff",
    offWhite: "#F5F5F5",
  },

  // Shadow colors (opacities)
  shadow: {
    subtle: "0 0.5px 2px rgba(0, 0, 0, 0.08)",
    default: "0 1px 3px rgba(0, 0, 0, 0.08)",
    medium: "0 4px 12px rgba(0, 0, 0, 0.12)",
    large: "0 8px 16px rgba(0, 0, 0, 0.15)",
  },

  // Border colors with opacity
  border: {
    subtle: "rgba(0, 0, 0, 0.05)",
    default: "#f1f2f4",
    muted: "#e5e7eb",
  },

  // Overlay colors
  overlay: {
    dark: "rgba(255, 255, 255, 0.95)",
    blur: "rgba(255, 255, 255, 0.9)",
  },

  // Status/validation colors
  status: {
    error: "#dc2626",
    errorBright: "#ef4444",
    success: "#16a34a",
    successAlt: "#10b981",
    warning: "#f59e0b",
    info: "#3b82f6",
  },

  // Button/interaction states
  button: {
    defaultBg: "rgba(0, 0, 0, 0.05)",
    hoverBg: "rgba(239, 68, 68, 0.1)",
    hoverColor: "#ef4444",
    disabledBg: "#E5E7EB",
    disabledText: "#9CA3AF",
  },

  // Category colors
  category: {
    default: "#8577ce",
  },

  // Fallback/special colors
  fallback: "#27b981",
} as const;

/**
 * Usage Examples:
 *
 * In TypeScript/React:
 * ```tsx
 * import { TRAYA_COLORS } from "@/theme/traya-colors";
 *
 * // Direct usage
 * style={{ background: TRAYA_COLORS.primary }}
 *
 * // With styled-system
 * style={{ background: TRAYA_COLORS.secondary }}
 *
 * // In conditional styling
 * const bgColor = isActive ? TRAYA_COLORS.primary : TRAYA_COLORS.secondary;
 * ```
 *
 * In Panda CSS recipes:
 * ```typescript
 * import { TRAYA_COLORS } from "@/theme/traya-colors";
 *
 * export const buttonRecipe = defineRecipe({
 *   base: {
 *     background: TRAYA_COLORS.primary,
 *     color: "white",
 *     _hover: {
 *       background: TRAYA_COLORS.secondary,
 *     },
 *   },
 * });
 * ```
 */
