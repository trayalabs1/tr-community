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
    // The page may scroll the document or an inner container depending on the
    // breakpoint, so read whichever is actually moving rather than assuming the
    // window. body carries overflow-y: clip here, which makes the scrolling
    // element differ between desktop and mobile.
    function currentY() {
      const docY = window.scrollY || document.documentElement.scrollTop || 0;
      if (docY > 0) return docY;
      const el = scrollerRef;
      return el ? el.scrollTop : 0;
    }

    // Remembered once found, so every frame is not a DOM search.
    let scrollerRef: HTMLElement | null = null;

    function findScroller(target: EventTarget | null) {
      if (!target || target === document || target === window) return null;
      const el = target as HTMLElement;
      if (typeof el.scrollTop !== "number") return null;
      return el.scrollHeight > el.clientHeight + 1 ? el : null;
    }

    let lastY = currentY();
    // Where the current run of travel in one direction began, so a slow scroll
    // accumulates toward the threshold instead of being discarded per event.
    let anchorY = lastY;
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
      const y = currentY();

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

    function onScroll(event: Event) {
      // Latch onto the inner container the first time one reports a scroll, so
      // the same hook works whether the document or a container is moving.
      if (!scrollerRef) {
        scrollerRef = findScroller(event.target);
      }
      // Scroll fires far more often than the screen repaints; coalescing into a
      // frame keeps this off the critical path during a fast flick.
      if (frame === 0) frame = window.requestAnimationFrame(read);
    }

    // Capture phase on document: scroll does not bubble, so a listener on
    // window alone misses an inner scroll container entirely.
    document.addEventListener("scroll", onScroll, {
      passive: true,
      capture: true,
    });
    return () => {
      document.removeEventListener("scroll", onScroll, { capture: true });
      if (frame !== 0) window.cancelAnimationFrame(frame);
    };
  }, []);

  return isScrolledDown;
}
