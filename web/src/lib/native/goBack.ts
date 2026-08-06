/**
 * Leave the community and return to the app.
 *
 * The shell watches the webview's URL and closes it when destination=go_back
 * appears. This is the same three lines as MobileCommandBar's handleCloseWebView,
 * behind the Home button — the established way out of the community.
 *
 * Deliberately unconditional, matching that helper: no bridge sniffing and no
 * history.length check. Next.js hydration pushes history entries, so a
 * history-based back would land on a blank state instead of leaving.
 */
export function closeWebView() {
  if (typeof window === "undefined") return;

  const url = new URL(window.location.href);
  url.searchParams.set("destination", "go_back");
  window.location.href = url.toString();
}
