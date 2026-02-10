export function generateRandomUsername(name?: string): string {
  const fourDigitSuffix = Math.floor(1000 + Math.random() * 9000);

  if (name) {
    const cleanName = name.toLowerCase().replace(/\s+/g, "").replace(/[^a-z0-9]/g, "");
    return `${cleanName}${fourDigitSuffix}`;
  }

  return `user${fourDigitSuffix}`;
}
