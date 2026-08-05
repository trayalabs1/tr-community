"use client";

import { Carousel } from "@ark-ui/react/carousel";
import { useState } from "react";

import { styled } from "@/styled-system/jsx";

type Props = {
  images: string[];
  /** Rounds the viewport corners. Feed cards want this; lightboxes may not. */
  rounded?: boolean;
};

// Slides sit in a fixed square and scale to fit inside it, so a 16:9 image keeps
// its shape and gains cream bands rather than being cropped. object-fit:
// contain is the CSS equivalent of the spec's scale = MIN(viewport / image) on
// both axes.
const SQUARE = "1 / 1";
// The theme's amber.light.2 — warm without the peach cast of the orange scale,
// which already signals "in review" on comment cards.
const BACKDROP = "#FEFBE9";

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
// to the square and gains cream bands left and right.
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

function ImageCarousel({
  images,
  rounded,
}: {
  images: string[];
  rounded: boolean;
}) {
  const [page, setPage] = useState(0);

  return (
    <Carousel.Root
      slideCount={images.length}
      page={page}
      onPageChange={(details) => setPage(details.page)}
      allowMouseDrag
      style={{ width: "100%" }}
    >
      <styled.div position="relative" width="full">
        <Carousel.ItemGroup
          style={{
            width: "100%",
            aspectRatio: SQUARE,
            borderRadius: rounded ? "var(--radii-lg)" : undefined,
            overflow: "hidden",
            // Grid tracks default to min-content, which a tall slide can grow
            // past. Pin the row so the square always wins.
            gridAutoRows: "100%",
            alignItems: "stretch",
          }}
        >
          {images.map((src, index) => (
            <Carousel.Item
              key={src}
              index={index}
              // A very tall image would otherwise stretch the grid track it
              // sits in — height: 100% resolves against that track, not the
              // square — so the slide is pinned to the viewport and clips.
              style={{ height: "100%", minHeight: 0, overflow: "hidden" }}
            >
              <styled.div
                width="full"
                height="full"
                display="flex"
                alignItems="center"
                justifyContent="center"
                style={{ backgroundColor: BACKDROP }}
              >
                <styled.img
                  src={src}
                  alt=""
                  style={{
                    maxWidth: "100%",
                    maxHeight: "100%",
                    width: "auto",
                    height: "auto",
                    objectFit: "contain",
                  }}
                />
              </styled.div>
            </Carousel.Item>
          ))}
        </Carousel.ItemGroup>

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

      <Carousel.IndicatorGroup
        style={{
          display: "flex",
          gap: "6px",
          justifyContent: "center",
          paddingTop: "8px",
        }}
      >
        {images.map((src, index) => (
          <Carousel.Indicator
            key={src}
            index={index}
            aria-label={`Go to image ${index + 1}`}
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
      </Carousel.IndicatorGroup>
    </Carousel.Root>
  );
}
