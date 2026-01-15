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
  // Primary brand color
  primary: "#329866",

  // Secondary color - used for backgrounds, tags, icons
  secondary: "#B2E6CD",

  // Tertiary color - light background, hover states
  tertiary: "#DCF4E8",

  // Gradient - primary to secondary
  gradient: "linear-gradient(180deg, #3EBC7F 17.69%, #329866 100%)",

  // Accent colors
  heart: "#F04343",

  // Neutral colors for complementary styling
  neutral: {
    light: "#F3F4F6",
    border: "#E5E7EB",
    text: "#6B7280",
    textMuted: "#9CA3AF",
  },
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
