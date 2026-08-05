declare global {
  interface Window {
    /** Injected by the React Native shell when it wants to own back navigation. */
    __rnGoBack?: () => void;
    ReactNativeWebView?: {
      postMessage: (message: string) => void;
    };
  }
}

/**
 * The app watches the webview's URL and closes it when this parameter appears.
 * See MobileCommandBar's Home button, which is the established way out of the
 * community back to the app.
 */
const DESTINATION_PARAM = "destination";
const GO_BACK_VALUE = "go_back";

/**
 * Leave the community and return to the app.
 *
 * The primary mechanism is the destination=go_back URL parameter, because that is
 * what the native shell actually listens for — the same route MobileCommandBar's
 * Home button takes. __rnGoBack still wins when the shell injects it, and
 * postMessage is sent alongside for shells that consume it.
 *
 * `fallback` runs only in a browser, or when the parameter is already set and the
 * shell has not acted, so a second press still does something.
 *
 * Deliberately does not consult history.length: Next.js hydration pushes history
 * entries, so it is always > 1 inside the webview and a history-based back lands
 * on a blank state instead of leaving. Screens that need to move within the
 * webview should keep using router.back().
 */
export function goBackToApp(fallback?: () => void) {
  if (typeof window === "undefined") return;

  if (window.__rnGoBack) {
    window.__rnGoBack();
    return;
  }

  const url = new URL(window.location.href);

  // Already asked to leave and the shell has not acted: fall back to in-app
  // navigation rather than re-navigating to the same URL, which would do nothing.
  if (url.searchParams.get(DESTINATION_PARAM) === GO_BACK_VALUE) {
    if (fallback) fallback();
    return;
  }

  // In a plain browser there is no shell to hand off to, so honour the caller's
  // in-app fallback instead of navigating with a parameter nothing will consume.
  if (!isEmbeddedInApp() && fallback) {
    fallback();
    return;
  }

  window.ReactNativeWebView?.postMessage("goBack");

  url.searchParams.set(DESTINATION_PARAM, GO_BACK_VALUE);
  window.location.href = url.toString();
}

/**
 * True when running inside the app's webview.
 *
 * The bridge globals are the reliable signal when present, but the shell does not
 * always inject them, so a webview user-agent counts too — that is the same
 * assumption MobileCommandBar makes by always offering its Home button.
 */
export function isEmbeddedInApp() {
  if (typeof window === "undefined") return false;
  if (window.__rnGoBack || window.ReactNativeWebView) return true;

  const ua = window.navigator.userAgent;
  // Android WebView reports "wv"; iOS WKWebView is a mobile Safari UA without
  // "Safari/", which a real Safari always carries.
  const isAndroidWebView = /\bwv\b/.test(ua) || /Version\/[\d.]+ Chrome/.test(ua);
  const isIOSWebView = /iPhone|iPad|iPod/.test(ua) && !/Safari\//.test(ua);
  return isAndroidWebView || isIOSWebView;
}
