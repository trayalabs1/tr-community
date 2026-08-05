"use client";

import { useEffect, useState } from "react";

/**
 * Asymmetric thresholds. Collapsing the header removes its height from the
 * document, which shifts scrollY and reads back as a small upward scroll — so a
 * symmetric threshold makes the two states fight each other and flicker during a
 * slow scroll. Requiring a larger deliberate movement to re-expand than the
 * collapse itself can produce breaks that loop.
 */
const COLLAPSE_AFTER_PX = 24;
const EXPAND_AFTER_PX = 64;
/** Keep everything visible near the top so the header never collapses at rest. */
const TOP_ZONE_PX = 32;
/**
 * A state change relayouts the page and emits more scroll events. Ignore them
 * for slightly longer than the CSS transition so the settling motion cannot be
 * mistaken for a new gesture.
 */
const SETTLE_MS = 260;

/**
 * True once the reader has scrolled down, false again as soon as they scroll up.
 * Used to collapse the parts of a sticky header that are not needed while
 * reading.
 */
export function useIsScrolledDown() {
  const [isScrolledDown, setScrolledDown] = useState(false);

  useEffect(() => {
    let lastY = window.scrollY;
    // Where the current run of travel in one direction began, so a slow scroll
    // accumulates toward the threshold instead of being discarded per event.
    let anchorY = window.scrollY;
    let collapsed = false;
    let frame = 0;
    let lockedUntil = 0;

    function apply(next: boolean) {
      if (next === collapsed) return;
      collapsed = next;
      lockedUntil = performance.now() + SETTLE_MS;
      setScrolledDown(next);
    }

    function read() {
      frame = 0;
      const y = window.scrollY;

      if (y <= TOP_ZONE_PX) {
        lastY = y;
        anchorY = y;
        apply(false);
        return;
      }

      // Still absorbing the relayout from the last change.
      if (performance.now() < lockedUntil) {
        lastY = y;
        anchorY = y;
        return;
      }

      // Direction reversed — start measuring the new run from here.
      if ((y > lastY && anchorY > lastY) || (y < lastY && anchorY < lastY)) {
        anchorY = lastY;
      }

      const travel = y - anchorY;
      lastY = y;

      if (!collapsed && travel > COLLAPSE_AFTER_PX) {
        apply(true);
        anchorY = y;
      } else if (collapsed && travel < -EXPAND_AFTER_PX) {
        apply(false);
        anchorY = y;
      }
    }

    function onScroll() {
      // Scroll fires far more often than the screen repaints; coalescing into a
      // frame keeps this off the critical path during a fast flick.
      if (frame === 0) frame = window.requestAnimationFrame(read);
    }

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => {
      window.removeEventListener("scroll", onScroll);
      if (frame !== 0) window.cancelAnimationFrame(frame);
    };
  }, []);

  return isScrolledDown;
}
