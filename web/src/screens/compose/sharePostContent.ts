export function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

export function isSafeMediaUrl(value: string): boolean {
  try {
    const { protocol } = new URL(value);
    return protocol === "http:" || protocol === "https:";
  } catch {
    return false;
  }
}

export function parseMediaParam(value: string | undefined): string[] {
  if (!value) return [];
  return value
    .split(",")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

export function buildGenericBody(text: string, media: string[]): string {
  const paragraph = `<p>${escapeHtml(text)}</p>`;
  const images = media
    .map((url) => `<img src="${escapeHtml(url)}" alt="" />`)
    .join("");
  return `${paragraph}${images}`;
}
