import imageCompression from "browser-image-compression";

/** Hard ceiling for an uploaded image. */
export const MAX_UPLOAD_BYTES = 2 * 1024 * 1024;

/**
 * Compression runs in a Web Worker (see useWebWorker below), which is the point:
 * decoding a phone photo allocates tens of megabytes, and doing that on the main
 * thread is what fails with "unable to complete previous operation due to low
 * memory" in the app's webview. This mirrors the settings the hair-test form uses,
 * which handles the same camera uploads on the same devices.
 */
const OPTIONS = {
  maxSizeMB: 1,
  maxWidthOrHeight: 1600,
  useWebWorker: true,
} as const;

function canCompress(file: File) {
  // GIFs would lose their animation and SVGs are vectors that would be
  // rasterised, so neither should be re-encoded.
  return /^image\/(jpeg|png|webp)$/i.test(file.type);
}

/**
 * Shrink an image to fit under MAX_UPLOAD_BYTES, preserving aspect ratio.
 *
 * Returns the original file when it already fits, when its format should not be
 * re-encoded, or when compression fails — a file this cannot process is not
 * necessarily one the server will reject.
 */
export async function compressImage(file: File): Promise<File> {
  if (!canCompress(file)) return file;
  if (file.size <= MAX_UPLOAD_BYTES) return file;

  try {
    const compressed = await imageCompression(file, OPTIONS);

    // Guard against a pathological case where the round-trip grows the file.
    if (compressed.size >= file.size) return file;

    return new File([compressed], file.name, {
      type: compressed.type || file.type,
      lastModified: file.lastModified,
    });
  } catch {
    return file;
  }
}
