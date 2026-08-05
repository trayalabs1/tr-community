"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { styled } from "@/styled-system/jsx";

type Props = {
  images: string[];
  /** Rounds the viewport corners. Feed cards want this; lightboxes may not. */
  rounded?: boolean;
};

// Slides sit in a fixed square and scale to fit inside it, so a 16:9 image keeps
// its shape and gains white bands rather than being cropped. object-fit:
// contain is the CSS equivalent of the spec's scale = MIN(viewport / image) on
// both axes.
const SQUARE = "1 / 1";
const BACKDROP = "#FFFFFF";

export function PostImages({ images, rounded = true }: Props) {
  if (images.length === 0) {
    return null;
  }

  if (images.length === 1) {
    return <SingleImage src={images[0]!} rounded={rounded} />;
  }

  return <ImageCarousel images={images} rounded={rounded} />;
}

// A single image keeps a dynamic height: it fills the width and grows as tall as
// its aspect ratio needs, up to a square. Anything taller than square is clamped
// to the square and gains white bands left and right.
function SingleImage({ src, rounded }: { src: string; rounded: boolean }) {
  const [ratio, setRatio] = useState<number | null>(null);

  // The ratio is unknown until the image loads, so reserve a square — the
  // tallest it can end up. The box then only ever shrinks to its final height
  // instead of pushing the rest of the post down.
  const isTallerThanSquare = ratio !== null && ratio < 1;
  const aspectRatio =
    ratio === null || isTallerThanSquare ? SQUARE : String(ratio);

  return (
    <styled.div
      width="full"
      overflow="hidden"
      borderRadius={rounded ? "lg" : "none"}
      style={{ aspectRatio, backgroundColor: BACKDROP }}
    >
      <styled.img
        src={src}
        alt=""
        width="full"
        height="full"
        style={{ objectFit: "contain" }}
        onLoad={(e) => {
          const img = e.currentTarget;
          if (img.naturalHeight > 0) {
            setRatio(img.naturalWidth / img.naturalHeight);
          }
        }}
      />
    </styled.div>
  );
}

// Built on native CSS scroll snapping rather than a carousel library: swiping
// is then just the browser scrolling the track, which no drag state machine can
// swallow. The page is derived from scroll position, and the dots scroll the
// track back — one source of truth either way.
function ImageCarousel({
  images,
  rounded,
}: {
  images: string[];
  rounded: boolean;
}) {
  const trackRef = useRef<HTMLDivElement>(null);
  const [page, setPage] = useState(0);

  const syncPage = useCallback(() => {
    const track = trackRef.current;
    if (!track || track.clientWidth === 0) return;
    const next = Math.round(track.scrollLeft / track.clientWidth);
    setPage(Math.min(Math.max(next, 0), images.length - 1));
  }, [images.length]);

  useEffect(() => {
    const track = trackRef.current;
    if (!track) return;
    track.addEventListener("scroll", syncPage, { passive: true });
    return () => track.removeEventListener("scroll", syncPage);
  }, [syncPage]);

  function goTo(index: number) {
    const track = trackRef.current;
    if (!track) return;
    track.scrollTo({ left: index * track.clientWidth, behavior: "smooth" });
  }

  return (
    <styled.div width="full">
      <styled.div position="relative" width="full">
        <styled.div
          ref={trackRef}
          className="postimages__track"
          display="flex"
          width="full"
          style={{
            aspectRatio: SQUARE,
            borderRadius: rounded ? "var(--radii-lg)" : undefined,
            overflowX: "auto",
            overflowY: "hidden",
            scrollSnapType: "x mandatory",
            overscrollBehaviorX: "contain",
            backgroundColor: BACKDROP,
          }}
        >
          {images.map((src) => (
            <styled.div
              key={src}
              flexShrink="0"
              width="full"
              height="full"
              display="flex"
              alignItems="center"
              justifyContent="center"
              style={{ scrollSnapAlign: "start", backgroundColor: BACKDROP }}
            >
              <styled.img
                src={src}
                alt=""
                // Native image drag would otherwise capture the pointer and
                // stop the track from scrolling under a mouse drag.
                draggable={false}
                style={{
                  maxWidth: "100%",
                  maxHeight: "100%",
                  width: "auto",
                  height: "auto",
                  objectFit: "contain",
                  userSelect: "none",
                  WebkitUserSelect: "none",
                }}
              />
            </styled.div>
          ))}
        </styled.div>

        <styled.div
          position="absolute"
          style={{
            top: "12px",
            right: "12px",
            padding: "2px 10px",
            borderRadius: "999px",
            backgroundColor: "rgba(0, 0, 0, 0.6)",
            color: "#FFFFFF",
            fontSize: "12px",
            lineHeight: "18px",
            fontWeight: 500,
            pointerEvents: "none",
          }}
        >
          {page + 1}/{images.length}
        </styled.div>
      </styled.div>

      <styled.div
        display="flex"
        justifyContent="center"
        style={{ gap: "6px", paddingTop: "8px" }}
      >
        {images.map((src, index) => (
          <styled.button
            key={src}
            type="button"
            aria-label={`Go to image ${index + 1}`}
            aria-current={index === page}
            onClick={() => goTo(index)}
            style={{
              width: index === page ? "16px" : "6px",
              height: "6px",
              borderRadius: "999px",
              border: "none",
              padding: 0,
              cursor: "pointer",
              transition: "width 0.2s ease, background-color 0.2s ease",
              backgroundColor: index === page ? "#404040" : "#D9D9D9",
            }}
          />
        ))}
      </styled.div>

      <style jsx global>{`
        .postimages__track {
          scrollbar-width: none;
          -ms-overflow-style: none;
          /* Horizontal gestures scroll the track; vertical ones stay with the
             page so the feed still scrolls under a swipe. */
          touch-action: pan-y;
        }
        .postimages__track::-webkit-scrollbar {
          display: none;
        }
      `}</style>
    </styled.div>
  );
}
