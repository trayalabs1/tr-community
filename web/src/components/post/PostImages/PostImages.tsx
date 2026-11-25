"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import {
  ChevronLeftIcon,
  ChevronRightIcon,
} from "@/components/ui/icons/Chevron";
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
// swallow. The page is derived from scroll position, and the arrows scroll the
// track — one source of truth either way, so a swipe and a tap cannot disagree.
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

    const target = Math.min(Math.max(index, 0), images.length - 1);
    const left = target * track.clientWidth;

    // Update state here rather than waiting for the scroll listener: smooth
    // scrolling can be suppressed (reduced-motion settings, some embedded
    // webviews) and the arrows must not stall when it is.
    setPage(target);

    // Assign directly and let CSS scroll-behavior animate it. scrollTo with
    // behavior: "smooth" is silently ignored in some embedded webviews, which
    // leaves the arrows doing nothing; a plain assignment always lands.
    track.scrollLeft = left;
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

        <ArrowButton
          side="left"
          disabled={page === 0}
          onClick={() => goTo(page - 1)}
        />
        <ArrowButton
          side="right"
          disabled={page === images.length - 1}
          onClick={() => goTo(page + 1)}
        />
      </styled.div>

      <style jsx global>{`
        .postimages__track {
          scrollbar-width: none;
          -ms-overflow-style: none;
          /* Horizontal gestures scroll the track; vertical ones stay with the
             page so the feed still scrolls under a swipe. */
          touch-action: pan-y;
          /* Animates the arrows' scrollLeft assignment. Declared in CSS rather
             than passed to scrollTo so it applies however the scroll is set. */
          scroll-behavior: smooth;
        }
        @media (prefers-reduced-motion: reduce) {
          .postimages__track {
            scroll-behavior: auto;
          }
        }
        .postimages__track::-webkit-scrollbar {
          display: none;
        }
      `}</style>
    </styled.div>
  );
}

/** Overlaid previous/next control. Disabled at the ends rather than looping, so
 *  the position in the set stays legible. */
function ArrowButton({
  side,
  disabled,
  onClick,
}: {
  side: "left" | "right";
  disabled: boolean;
  onClick: () => void;
}) {
  const Icon = side === "left" ? ChevronLeftIcon : ChevronRightIcon;

  return (
    <styled.button
      type="button"
      aria-label={side === "left" ? "Previous image" : "Next image"}
      disabled={disabled}
      onClick={onClick}
      position="absolute"
      display="flex"
      alignItems="center"
      justifyContent="center"
      style={{
        top: "50%",
        transform: "translateY(-50%)",
        [side]: "8px",
        width: "32px",
        height: "32px",
        borderRadius: "999px",
        border: "none",
        padding: 0,
        cursor: disabled ? "default" : "pointer",
        backgroundColor: "rgba(0, 0, 0, 0.55)",
        color: "#FFFFFF",
        // Hidden rather than unmounted at the ends, so the remaining arrow does
        // not shift and the hit area stays where the reader last tapped.
        opacity: disabled ? 0 : 1,
        pointerEvents: disabled ? "none" : "auto",
        transition: "opacity 0.15s ease",
      }}
    >
      <Icon width="5" height="5" />
    </styled.button>
  );
}
