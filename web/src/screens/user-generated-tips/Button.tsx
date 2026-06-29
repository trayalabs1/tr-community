'use client';

/**
 * Button — see `tatva-design-system/components/atoms/button.md`.
 *
 * Cross-surface atom. STRICT visual parity with
 * `trayadesign-app/components/Button.tsx`. Variant / size / state matrix is
 * identical; only the API differs where the underlying primitive differs:
 *   - `onClick` (web) ↔ `onPress` (RN)
 *   - `icon: ComponentType<{...}>` accepts any `lucide-react` icon component
 *
 * Sourced from Figma node 17408:19849 (Buttons).
 */

import React, { CSSProperties, useState } from 'react';

export type ButtonVariant =
  | 'primary'
  | 'dark'
  | 'outline'
  | 'muted'
  | 'blue'
  | 'white'
  | 'noFill';

export type ButtonSize = 'xl' | 'lg' | 'md' | 'sm';

// Minimal shape any lucide-react icon component satisfies.
export type IconComponent = React.ComponentType<{
  size?: number;
  color?: string;
  strokeWidth?: number;
}>;

type Props = {
  label?: string;
  variant?: ButtonVariant;
  size?: ButtonSize;
  disabled?: boolean;
  loading?: boolean;
  icon?: IconComponent;
  iconPosition?: 'left' | 'right';
  iconOnly?: boolean;
  fullWidth?: boolean;
  onClick?: () => void;
};

type SizeSpec = {
  height: number;
  paddingH: number;
  radius: number;
  iconSize: number;
  gap: number;
  fontSize: number;
  lineHeight: number;
};

// Exact mirror of trayadesign-app/components/Button.tsx → SIZE_SPECS.
const SIZE_SPECS: Record<ButtonSize, SizeSpec> = {
  xl: { height: 48, paddingH: 24, radius: 12, iconSize: 20, gap: 8, fontSize: 16, lineHeight: 20 },
  lg: { height: 40, paddingH: 20, radius: 10, iconSize: 18, gap: 8, fontSize: 14, lineHeight: 20 },
  md: { height: 32, paddingH: 16, radius: 8,  iconSize: 16, gap: 6, fontSize: 14, lineHeight: 20 },
  sm: { height: 24, paddingH: 12, radius: 6,  iconSize: 14, gap: 4, fontSize: 12, lineHeight: 16 },
};

type VariantSpec = {
  bg: string;
  bgDisabled: string;
  fg: string;
  fgDisabled: string;
  border?: string;
  borderDisabled?: string;
  underline?: boolean;
};

// Exact mirror of trayadesign-app/components/Button.tsx → VARIANT_SPECS.
// Uses CSS vars (auto-generated from canonical tokens — same hex values as RN).
const VARIANT_SPECS: Record<ButtonVariant, VariantSpec> = {
  primary: {
    bg: 'var(--color-lime-500)',
    bgDisabled: 'var(--color-lime-200)',
    fg: 'var(--color-grey-700)',
    fgDisabled: 'var(--color-grey-300)',
  },
  dark: {
    bg: 'var(--color-grey-700)',
    bgDisabled: 'var(--color-grey-200)',
    fg: 'var(--color-grey-0)',
    fgDisabled: 'var(--color-grey-300)',
  },
  outline: {
    bg: 'transparent',
    bgDisabled: 'transparent',
    fg: 'var(--color-grey-700)',
    fgDisabled: 'var(--color-grey-300)',
    border: 'var(--color-grey-700)',
    borderDisabled: 'var(--color-grey-200)',
  },
  muted: {
    bg: 'var(--color-grey-100)',
    bgDisabled: 'var(--color-grey-50)',
    fg: 'var(--color-grey-700)',
    fgDisabled: 'var(--color-grey-300)',
  },
  blue: {
    bg: 'var(--color-blue-200)',
    bgDisabled: 'var(--color-blue-100)',
    fg: 'var(--color-blue-900)',
    fgDisabled: 'var(--color-blue-400)',
  },
  white: {
    bg: 'var(--color-grey-0)',
    bgDisabled: 'var(--color-grey-0)',
    fg: 'var(--color-grey-700)',
    fgDisabled: 'var(--color-grey-300)',
  },
  noFill: {
    bg: 'transparent',
    bgDisabled: 'transparent',
    fg: 'var(--color-grey-700)',
    fgDisabled: 'var(--color-grey-300)',
    underline: true,
  },
};

const FONT = "var(--font-nunito-sans), 'Nunito Sans', sans-serif";

/** Loading spinner — matches RN ActivityIndicator visual weight. */
function Spinner({ color, size }: { color: string; size: number }) {
  return (
    <div
      style={{
        width: size,
        height: size,
        border: `2px solid ${color}`,
        borderTopColor: 'transparent',
        borderRadius: '50%',
        animation: 'spin 0.7s linear infinite',
        boxSizing: 'border-box',
      }}
    />
  );
}

export default function Button({
  label,
  variant = 'primary',
  size = 'lg',
  disabled = false,
  loading = false,
  icon: Icon,
  iconPosition = 'left',
  iconOnly = false,
  fullWidth = false,
  onClick,
}: Props) {
  const spec = SIZE_SPECS[size];
  const v = VARIANT_SPECS[variant];
  const isDisabled = disabled || loading;

  // JS-tracked press state — mirrors RN's `Pressable.pressed`.
  // CSS `:active` doesn't fire reliably on iOS WebView for fast taps because
  // the active state clears before the transition can render. Tracking via
  // pointer events guarantees the scale/opacity feedback shows on every press.
  const [pressed, setPressed] = useState(false);
  const startPress = isDisabled ? undefined : () => setPressed(true);
  const endPress = isDisabled ? undefined : () => setPressed(false);

  const fg = isDisabled ? v.fgDisabled : v.fg;
  const bg = isDisabled ? v.bgDisabled : v.bg;
  const border = v.border ? (isDisabled ? v.borderDisabled : v.border) : undefined;

  const containerStyle: CSSProperties = {
    height: spec.height,
    // iconOnly = square; otherwise paddingH controls width and content controls intrinsic width.
    paddingLeft: iconOnly ? 0 : spec.paddingH,
    paddingRight: iconOnly ? 0 : spec.paddingH,
    paddingTop: 0,
    paddingBottom: 0,
    width: iconOnly ? spec.height : (fullWidth ? '100%' : undefined),
    borderRadius: variant === 'noFill' ? 0 : spec.radius,
    backgroundColor: bg,
    border: border ? `1px solid ${border}` : 'none',
    color: fg,
    fontFamily: FONT,
    fontSize: spec.fontSize,
    fontWeight: 600,
    lineHeight: `${spec.lineHeight}px`,
    textDecoration: v.underline ? 'underline' : 'none',
    display: 'inline-flex',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spec.gap,
    cursor: isDisabled ? 'default' : 'pointer',
    flexShrink: 0,
    boxSizing: 'border-box',
    whiteSpace: 'nowrap',
    // Mirror RN's Pressable.pressed feedback: transform: scale(0.97) + opacity: 0.95.
    transform: pressed ? 'scale(0.97)' : 'scale(1)',
    opacity: pressed ? 0.95 : 1,
    transition: 'transform 0.08s ease, opacity 0.08s ease',
  };

  return (
    <button
      className="tatva-button"
      onClick={isDisabled ? undefined : onClick}
      // Press state via pointer events (works for touch + mouse + stylus uniformly).
      // onPointerCancel/onPointerLeave clear it for swipe-off / multi-touch / drag-away.
      onPointerDown={startPress}
      onPointerUp={endPress}
      onPointerCancel={endPress}
      onPointerLeave={endPress}
      // Empty onTouchStart — additional hint to iOS WebView that the element is interactive.
      onTouchStart={() => {}}
      disabled={isDisabled}
      style={containerStyle}
    >
      {loading ? (
        <Spinner color={fg} size={spec.iconSize} />
      ) : (
        <>
          {Icon && (iconPosition === 'left' || iconOnly) && (
            <Icon size={spec.iconSize} color={fg} strokeWidth={2} />
          )}
          {!iconOnly && label !== undefined && <span>{label}</span>}
          {Icon && iconPosition === 'right' && !iconOnly && (
            <Icon size={spec.iconSize} color={fg} strokeWidth={2} />
          )}
        </>
      )}
    </button>
  );
}
