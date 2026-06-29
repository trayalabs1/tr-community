'use client';

function goBack() {
  if (typeof window === 'undefined') return;
  // Inside RN WebView — always let RN handle navigation.
  // Next.js hydration pushes history states so history.length is always > 1;
  // checking it here would call window.history.back() into a blank state.
  // Pages that need within-WebView back (e.g. a multi-step flow) should pass
  // an explicit onBack prop instead.
  if ((window as any).__rnGoBack) {
    (window as any).__rnGoBack();
  } else if ((window as any).ReactNativeWebView) {
    (window as any).ReactNativeWebView.postMessage('goBack');
  } else {
    window.history.back();
  }
}

/** A single right-edge icon action. The icon is rendered inside a 48×48
 *  hit-area button; the button carries the a11y label, the icon is decorative.
 *  For icon-group coherence the icons MUST share stroke width (2) and color
 *  with the bar's back arrow — see iconography/icons.md "Icon groups". */
export interface SubpageTopBarAction {
  icon: React.ReactNode;
  onClick: () => void;
  label: string;
  /** Optional notification count shown as a red badge on the icon's top-right.
   *  Numbers above 99 render as "99+". Falsy / 0 / '' hides the badge. Sized to
   *  fit up to two digits. */
  badge?: number | string;
}

const BADGE_FONT = "var(--font-nunito-sans), 'Nunito Sans', sans-serif";

function badgeText(badge: number | string): string {
  if (typeof badge === 'number') return badge > 99 ? '99+' : String(badge);
  return badge;
}

interface SubpageTopBarProps {
  /** Optional centered title. Omit for back-arrow-only top bars (e.g. flows
   *  whose page-level heading sits below the bar in the scroll body). */
  title?: string;
  onBack?: () => void;
  /** Hide the left back button (renders a spacer so the title stays centered).
   *  Use for flows that only exit via the close (X), never step back. */
  showBack?: boolean;
  /** Suppress the back arrow on the left (renders a spacer so the title stays
   *  centred). Use for modal-style screens that only carry a close (X) action.
   *  Mirrors the RN peer's `hideBack`. */
  hideBack?: boolean;
  /** When provided, renders an X close button on the right. Distinct from
   *  back: close always exits the entire flow (e.g. multi-step onboarding),
   *  while back returns to the previous step within it. */
  onClose?: () => void;
  /** Up to two right-edge icon actions (e.g. Report + Like on a profile).
   *  Each renders in a 48×48 hit area. Takes precedence over `rightSlot`
   *  and `onClose`. Capped at 2 — extra entries are ignored. */
  rightActions?: SubpageTopBarAction[];
  /** Custom free-form right-slot content. Takes precedence over onClose.
   *  Prefer `rightActions` for icon buttons — it guarantees hit-area + a11y. */
  rightSlot?: React.ReactNode;
  /** Optional solid fill colour. When provided, paints a solid block (no
   *  frosted-glass blur) — extending into the safe-area top region so the OS
   *  status bar reads the same colour. Mirrors the RN peer's `backgroundColor`.
   *  Omit for the default frosted-glass look. */
  backgroundColor?: string;
}

export default function SubpageTopBar({ title, onBack, onClose, hideBack = false, rightActions, rightSlot, backgroundColor }: SubpageTopBarProps) {
  const handleBack = onBack ?? goBack;
  const actions = rightActions?.slice(0, 2);
  const solid = backgroundColor != null;

  // Bar uses `position: fixed` (not `sticky`) so it stays pinned to the viewport
  // even when the keyboard opens. iOS WebView's outer scroll view scrolls the
  // entire page up to keep the focused input visible — a sticky bar gets pulled
  // up with it; a fixed bar stays put. An in-flow spacer below this element
  // reserves the same vertical space so subsequent content doesn't sit underneath.
  return (
    <>
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      zIndex: 10,
      height: 'calc(56px + var(--rn-sat, env(safe-area-inset-top)))',
      paddingTop: 'var(--rn-sat, env(safe-area-inset-top))',
      display: 'flex',
      alignItems: 'center',
      paddingLeft: 4,
      paddingRight: 4,
      backgroundColor: solid ? backgroundColor : 'rgba(255,255,255,0.7)',
      backdropFilter: solid ? undefined : 'blur(12px)',
      transition: 'background-color 260ms ease',
      flexShrink: 0,
    }}>
      {/* Back button — or a 48×48 spacer when hidden, keeping the title centred */}
      {hideBack ? (
        <div style={{ width: 48, height: 48, flexShrink: 0 }} />
      ) : (
      <button
        onClick={handleBack}
        aria-label="Go back"
        style={{
          width: 48,
          height: 48,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          padding: 0,
          flexShrink: 0,
        }}
      >
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none"
          stroke="#2C2C2A" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <line x1="19" y1="12" x2="5" y2="12" />
          <polyline points="12 19 5 12 12 5" />
        </svg>
      </button>
      )}

      {/* Centered title (omitted when `title` is not provided — leaves a flex
          spacer so the right slot stays at the far edge). */}
      <div style={{
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
      }}>
        {title ? (
          <p style={{
            margin: 0,
            fontSize: 20,
            fontWeight: 600,
            lineHeight: '24px',
            letterSpacing: '-0.1px',
            color: 'var(--color-grey-700)',
            fontFamily: "var(--font-nunito-sans), 'Nunito Sans', sans-serif",
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            textAlign: 'center',
          }}>
            {title}
          </p>
        ) : null}
      </div>

      {/* Right slot — icon actions, free-form content, close button, or spacer.
          Precedence: rightActions > rightSlot > onClose > spacer (keeps title centred). */}
      {actions && actions.length > 0 ? (
        <div style={{ display: 'flex', alignItems: 'center', paddingRight: 4, flexShrink: 0 }}>
          {actions.map((action, i) => (
            <button
              key={i}
              type="button"
              onClick={action.onClick}
              aria-label={action.label}
              style={{
                width: 48,
                height: 48,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: 0,
                flexShrink: 0,
              }}
            >
              <span style={{ position: 'relative', display: 'inline-flex' }}>
                {action.icon}
                {action.badge != null && action.badge !== '' && action.badge !== 0 && (
                  <span
                    aria-hidden
                    style={{
                      position: 'absolute',
                      top: -7,
                      right: -5,
                      minWidth: 18,
                      height: 18,
                      borderRadius: 9,
                      backgroundColor: 'var(--color-signal-warning-300)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      padding: '0 4px',
                      boxSizing: 'border-box',
                      fontFamily: BADGE_FONT,
                      fontWeight: 600,
                      fontSize: 10,
                      lineHeight: '12px',
                      color: 'var(--color-grey-0)',
                    }}
                  >
                    {badgeText(action.badge)}
                  </span>
                )}
              </span>
            </button>
          ))}
        </div>
      ) : rightSlot ? (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, paddingRight: 12, flexShrink: 0 }}>
          {rightSlot}
        </div>
      ) : onClose ? (
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          style={{
            width: 48,
            height: 48,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
            flexShrink: 0,
          }}
        >
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none"
            stroke="#2C2C2A" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      ) : (
        <div style={{ width: 48, height: 48, flexShrink: 0 }} />
      )}
    </div>
    {/* In-flow spacer so the (now-fixed) bar doesn't overlap page content. */}
    <div
      aria-hidden
      style={{
        height: 'calc(56px + var(--rn-sat, env(safe-area-inset-top)))',
        flexShrink: 0,
      }}
    />
    </>
  );
}
