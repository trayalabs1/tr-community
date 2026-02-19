export function generateRandomUsername(name?: string): string {
  const suffix = Math.floor(100 + Math.random() * 900);

  if (name) {
    const cleanName = name
      .toLowerCase()
      .replace(/\s+/g, "")
      .replace(/[^a-z0-9]/g, "");

    return `${shuffleString(cleanName)}${suffix}`;
  }

  return `user${suffix}`;
}

function shuffleString(str: string): string {
  const arr = str.split("");
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    const temp = arr[i];
    arr[i] = arr[j]!;
    arr[j] = temp!;
  }
  return arr.join("");
}
