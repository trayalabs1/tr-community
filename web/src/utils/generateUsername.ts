export function generateRandomUsername(name?: string): string {
  const suffix = Math.floor(100 + Math.random() * 900);

  if (name) {
    const cleanName = name
      .toLowerCase()
      .replace(/\s+/g, "")
      .replace(/[^a-z0-9]/g, "");

    const prefix = cleanName.slice(0, 3) || "user";
    return `${prefix}${suffix}`;
  }

  return `user${suffix}`;
}
