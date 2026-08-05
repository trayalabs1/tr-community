/** Hard ceiling for an uploaded image. */
export const MAX_UPLOAD_BYTES = 2 * 1024 * 1024;

/**
 * Longest edge to allow. Comfortably above the square the feed renders images
 * into, so downscaling to this is not visible in normal use.
 */
const MAX_EDGE = 1600;

/** Tried in order until the encoded blob fits under the ceiling. */
const QUALITY_STEPS = [0.85, 0.75, 0.65, 0.55, 0.45];

/**
 * Last resort: if even the lowest quality is too large the image is enormous, so
 * halve the edge and try the ladder again. Two passes covers a 100MP source.
 */
const MAX_DOWNSCALE_PASSES = 2;

function canCompress(file: File) {
  // GIFs would lose their animation, and SVGs are vectors that re-encode to a
  // raster. Neither belongs in a canvas round-trip.
  return /^image\/(jpeg|png|webp)$/i.test(file.type);
}

async function loadBitmap(file: File) {
  if (typeof createImageBitmap === "function") {
    return createImageBitmap(file);
  }

  // Safari below 15 has no createImageBitmap for blobs.
  const url = URL.createObjectURL(file);
  try {
    const img = new Image();
    img.src = url;
    await img.decode();
    return img;
  } finally {
    URL.revokeObjectURL(url);
  }
}

function encode(
  source: ImageBitmap | HTMLImageElement,
  width: number,
  height: number,
  quality: number,
): Promise<Blob | null> {
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;

  const context = canvas.getContext("2d");
  if (!context) return Promise.resolve(null);
  context.drawImage(source, 0, 0, width, height);

  return new Promise((resolve) => {
    canvas.toBlob(
      (blob) => {
        // Release the backing store immediately; on a memory-constrained webview
        // it is usually the largest single allocation in play.
        canvas.width = 0;
        canvas.height = 0;
        resolve(blob);
      },
      "image/jpeg",
      quality,
    );
  });
}

/**
 * Shrink an image until it fits under MAX_UPLOAD_BYTES, preserving aspect ratio.
 *
 * Quality is stepped down before resolution, so a photo keeps its dimensions and
 * loses only compression fidelity unless it is genuinely huge. Returns the
 * original file unchanged when it already fits, or when it is a format a canvas
 * round-trip would damage.
 */
export async function compressImage(file: File): Promise<File> {
  if (!canCompress(file)) return file;
  if (file.size <= MAX_UPLOAD_BYTES) return file;

  let source: ImageBitmap | HTMLImageElement;
  try {
    source = await loadBitmap(file);
  } catch {
    // Undecodable here does not mean the server will reject it.
    return file;
  }

  try {
    const sourceWidth = "width" in source ? source.width : 0;
    const sourceHeight = "height" in source ? source.height : 0;
    if (!sourceWidth || !sourceHeight) return file;

    let scale = Math.min(1, MAX_EDGE / Math.max(sourceWidth, sourceHeight));

    for (let pass = 0; pass <= MAX_DOWNSCALE_PASSES; pass++) {
      const width = Math.max(1, Math.round(sourceWidth * scale));
      const height = Math.max(1, Math.round(sourceHeight * scale));

      for (const quality of QUALITY_STEPS) {
        const blob = await encode(source, width, height, quality);
        if (!blob) return file;

        if (blob.size <= MAX_UPLOAD_BYTES) {
          return new File([blob], toJpegName(file.name), {
            type: "image/jpeg",
            lastModified: file.lastModified,
          });
        }
      }

      scale *= 0.5;
    }

    return file;
  } finally {
    if ("close" in source) source.close();
  }
}

function toJpegName(name: string) {
  const base = name.replace(/\.[^.]+$/, "") || "image";
  return `${base}.jpg`;
}
