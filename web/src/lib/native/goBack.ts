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
 * Leave the current screen, handing control back to the React Native shell when
 * the community is embedded in the app.
 *
 * Deliberately does not consult history.length: Next.js hydration pushes
 * history entries, so it is always > 1 inside the WebView and a history-based
 * back lands on a blank state instead of closing the webview. Screens that need
 * to move within the WebView should keep using router.back() instead.
 */
export function goBackToApp(fallback?: () => void) {
  if (typeof window === "undefined") return;

  if (window.__rnGoBack) {
    window.__rnGoBack();
    return;
  }

  if (window.ReactNativeWebView) {
    window.ReactNativeWebView.postMessage("goBack");
    return;
  }

  if (fallback) {
    fallback();
    return;
  }

  window.history.back();
}

/** True when running inside the React Native shell. */
export function isEmbeddedInApp() {
  if (typeof window === "undefined") return false;
  return Boolean(window.__rnGoBack || window.ReactNativeWebView);
}
