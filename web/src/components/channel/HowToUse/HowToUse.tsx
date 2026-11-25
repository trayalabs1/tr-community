"use client";

import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import {
  Bell,
  Bookmark,
  CircleHelp,
  Filter,
  Heart,
  MessageSquare,
  Search,
  Send,
  Share2,
  X,
} from "lucide-react";

import { accountUpdate, useAccountGet } from "src/api/openapi-client/accounts";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { styled } from "@/styled-system/jsx";

import { Account } from "@/api/openapi-schema";

import { InfoIcon } from "./InfoIcon";
import { MiniProfileAvatar } from "./MiniProfileAvatar";
import { PointerIcon } from "./PointerIcon";
import {
  HOW_TO_USE_SEEN_META_KEY,
  isNewCustomer,
  HOW_TO_USE_STEPS,
  HOW_TO_USE_TOOLTIP_MS,
  type HowToUseStep,
} from "./steps";

const ICONS = {
  search: Search,
  bell: Bell,
  send: Send,
  help: CircleHelp,
  filter: Filter,
  posts: Send,
  bookmark: Bookmark,
  like: Heart,
  comment: MessageSquare,
  share: Share2,
};

const PADDING = 6;
const REVEAL_BAND = 1100;
const REVEAL_SOLID = 880;
const REVEAL_MS = 600;
const REVEAL_WIPE_MS = 480;
const REVEAL_MASK = `linear-gradient(180deg, #000 0, #000 ${REVEAL_SOLID}px, transparent ${REVEAL_BAND}px)`;
const CARD_GAP = 59;
const VIEWPORT_MARGIN = 16;
const RADIUS = 16;

function cutoutPath(rect: Rect, pad: { x: number; y: number }) {
  const x = rect.left - pad.x;
  const y = rect.top - pad.y;
  const w = rect.width + pad.x * 2;
  const h = rect.height + pad.y * 2;
  const r = Math.min(RADIUS, w / 2, h / 2);
  const vw = window.innerWidth;
  const vh = window.innerHeight;
  return [
    `M0 0H${vw}V${vh}H0Z`,
    `M${x + r} ${y}H${x + w - r}A${r} ${r} 0 0 1 ${x + w} ${y + r}V${y + h - r}A${r} ${r} 0 0 1 ${x + w - r} ${y + h}H${x + r}A${r} ${r} 0 0 1 ${x} ${y + h - r}V${y + r}A${r} ${r} 0 0 1 ${x + r} ${y}Z`,
  ].join("");
}

const EXPAND_SETTLE_MS = 320;

function scrollToTop() {
  let scrolled = false;
  if ((window.scrollY || document.documentElement.scrollTop) > 0) {
    scrolled = true;
    window.scrollTo(0, 0);
    document.documentElement.scrollTop = 0;
    document.body.scrollTop = 0;
  }
  let el = document.querySelector<HTMLElement>('[data-htu="composer"]');
  while (el) {
    if (el.scrollTop > 0) {
      scrolled = true;
      el.scrollTop = 0;
    }
    el = el.parentElement;
  }
  return scrolled;
}

type Rect = { top: number; left: number; width: number; height: number };

function measure(target: string): Rect | null {
  const el = document.querySelector<HTMLElement>(`[data-htu="${target}"]`);
  if (!el) return null;
  const r = el.getBoundingClientRect();
  return { top: r.top, left: r.left, width: r.width, height: r.height };
}

function Spotlight({
  account,
  step,
  index,
  total,
  onNext,
  onDismiss,
}: {
  account?: Account;
  step: HowToUseStep;
  index: number;
  total: number;
  onNext: () => void;
  onDismiss: () => void;
}) {
  const [rect, setRect] = useState<Rect | null>(null);
  const [cardHeight, setCardHeight] = useState(0);
  const cardRef = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    const update = () => setRect(measure(step.target));
    update();
    window.addEventListener("resize", update);
    window.addEventListener("scroll", update, true);
    return () => {
      window.removeEventListener("resize", update);
      window.removeEventListener("scroll", update, true);
    };
  }, [step.target]);

  const dialogRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    const block = (event: Event) => event.preventDefault();
    el.addEventListener("wheel", block, { passive: false });
    el.addEventListener("touchmove", block, { passive: false });
    return () => {
      el.removeEventListener("wheel", block);
      el.removeEventListener("touchmove", block);
    };
  }, [rect !== null]);

  useEffect(() => {
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = prev;
    };
  }, []);

  const hasRect = rect !== null;
  useLayoutEffect(() => {
    const el = cardRef.current;
    if (!el) return;
    const sync = () => setCardHeight(el.getBoundingClientRect().height);
    sync();
    const observer = new ResizeObserver(sync);
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasRect]);

  if (!rect) return null;

  const last = index === total - 1;
  const pad = step.padding ?? { x: PADDING, y: PADDING };
  const viewportHeight = window.innerHeight;
  const below = rect.top + rect.height + pad.y + CARD_GAP;
  const above = rect.top - pad.y - CARD_GAP - cardHeight;
  const fitsBelow = below + cardHeight <= viewportHeight - VIEWPORT_MARGIN;
  const fitsAbove = above >= VIEWPORT_MARGIN;
  const cardTop = fitsBelow
    ? below
    : fitsAbove
      ? above
      : Math.max(VIEWPORT_MARGIN, viewportHeight - cardHeight - VIEWPORT_MARGIN);

  return createPortal(
    <styled.div
      ref={dialogRef}
      position="fixed"
      inset="0"
      style={{ zIndex: 10000 }}
      role="dialog"
      aria-label="How to use community"
    >
      <styled.div
        position="absolute"
        inset="0"
        onClick={onDismiss}
        style={
          index === 0
            ? {
                clipPath: "inset(0 0 100% 0)",
                animation: `htuWipe ${REVEAL_WIPE_MS}ms forwards`,
              }
            : undefined
        }
      >
        <styled.div
          position="absolute"
          inset="0"
          style={{
            backdropFilter: "blur(3.5px)",
            WebkitBackdropFilter: "blur(3.5px)",
            clipPath: `path(evenodd, "${cutoutPath(rect, pad)}")`,
            transition: "clip-path 200ms ease",
          }}
        />
        <styled.div
          position="absolute"
          inset="0"
          style={{
            background:
              "linear-gradient(180deg, rgba(0,0,0,0.81) 44.4169%, rgba(0,0,0,0.36) 88.6508%, rgba(0,0,0,0.27) 100%)",
            clipPath: `path(evenodd, "${cutoutPath(rect, pad)}")`,
            transition: "clip-path 200ms ease",
            ...(index === 0
              ? {
                  maskImage: REVEAL_MASK,
                  WebkitMaskImage: REVEAL_MASK,
                  maskSize: `100% ${REVEAL_BAND}px`,
                  WebkitMaskSize: `100% ${REVEAL_BAND}px`,
                  maskRepeat: "no-repeat",
                  WebkitMaskRepeat: "no-repeat",
                  ["--htu-end" as string]: `${viewportHeight - REVEAL_SOLID}px`,
                  animation: `htuReveal ${REVEAL_MS}ms forwards, htuFade ${REVEAL_MS}ms forwards`,
                }
              : {}),
          }}
        />
      </styled.div>
      <styled.div
        position="absolute"
        style={{
          top: rect.top - pad.y,
          left: rect.left - pad.x,
          width: rect.width + pad.x * 2,
          height: rect.height + pad.y * 2,
          borderRadius: RADIUS,
          boxShadow: "0 4px 5px rgba(0,0,0,0.25)",
          pointerEvents: "none",
        }}
      />
      <styled.div
        ref={cardRef}
        position="absolute"
        style={{
          top: cardTop,
          left: 28,
          right: 28,
          maxWidth: 568,
          margin: "0 auto",
          color: "#ffffff",
        }}
      >
        <styled.div
          style={{
            display: fitsBelow ? "block" : "none",
            position: "absolute",
            top: -44,
            left: 29,
            animation: step.showPointer
              ? "htuTap 1.2s ease-in-out infinite"
              : undefined,
          }}
        >
          <PointerIcon />
        </styled.div>
        <styled.div
          style={{
            display: "flex",
            alignItems: "flex-end",
            justifyContent: "space-between",
            fontFamily: "inherit",
          }}
        >
          <styled.div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 15,
              width: 243,
              maxWidth: "calc(100% - 92px)",
            }}
          >
            <styled.p
              style={{
                margin: 0,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
                fontSize: 18,
                lineHeight: "22px",
                fontWeight: 700,
                color: "#ffffff",
              }}
            >
              {step.title}
            </styled.p>
            <styled.ul
              style={{
                listStyle: "none",
                margin: 0,
                padding: 0,
                display: "flex",
                flexDirection: "column",
                gap: 14,
              }}
            >
              {step.items.map((item) => {
                const Icon = item.icon === "profile" ? null : ICONS[item.icon];
                return (
                  <styled.li
                    key={item.label}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      fontSize: 14,
                      lineHeight: "20px",
                      fontWeight: 600,
                      color: "#f7f7f7",
                    }}
                  >
                    <styled.span
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        justifyContent: "center",
                        width: 18,
                        height: 18,
                        flexShrink: 0,
                      }}
                    >
                      {Icon ? (
                        <Icon size={18} strokeWidth={1.75} />
                      ) : (
                        <MiniProfileAvatar account={account} />
                      )}
                    </styled.span>
                    {item.label}
                  </styled.li>
                );
              })}
            </styled.ul>
          </styled.div>
          <styled.div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "flex-end",
              justifyContent: "space-between",
              alignSelf: "stretch",
            }}
          >
            <styled.button
              type="button"
              aria-label="Close"
              onClick={onDismiss}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                width: 32,
                height: 32,
                padding: 6,
                background: "transparent",
                border: "none",
                color: "#ffffff",
                cursor: "pointer",
              }}
            >
              <X size={18} />
            </styled.button>
            <styled.button
              type="button"
              onClick={onNext}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                height: 32,
                padding: "0 20px",
                background: "#daebf1",
                color: "#2b5f73",
                border: "none",
                borderRadius: 8,
                fontFamily: "inherit",
                fontSize: 14,
                lineHeight: "20px",
                fontWeight: 600,
                whiteSpace: "nowrap",
                cursor: "pointer",
              }}
            >
              {last ? "Got it" : "Next"}
            </styled.button>
          </styled.div>
        </styled.div>
      </styled.div>
      <style>{`@keyframes htuTap{0%,100%{transform:translateY(0) scale(1)}50%{transform:translateY(-6px) scale(0.9)}}@keyframes htuReveal{0%{mask-position:0 -1100px;-webkit-mask-position:0 -1100px;animation-timing-function:cubic-bezier(.45,0,.2,1)}100%{mask-position:0 var(--htu-end);-webkit-mask-position:0 var(--htu-end)}}@keyframes htuWipe{0%{clip-path:inset(0 0 100% 0);animation-timing-function:cubic-bezier(.45,0,.2,1)}100%{clip-path:inset(0 0 0 0)}}@keyframes htuFade{0%{opacity:0}20%,100%{opacity:1}}`}</style>
    </styled.div>,
    document.body,
  );
}

export function HowToUseButton({ account }: { account?: Account }) {
  const {
    trackHtuTooltipShown,
    trackHtuClicked,
    trackHtuNextClicked,
    trackHtuDismissed,
  } = useEventTracking();
  const { data: loadedAccount, mutate: mutateAccount } = useAccountGet();
  const persisted = useRef(false);
  const [tooltip, setTooltip] = useState(false);
  const [phase, setPhase] = useState<number | null>(null);

  useEffect(() => {
    if (!loadedAccount) return;
    if (!isNewCustomer((loadedAccount as unknown as { joined?: unknown }).joined)) {
      return;
    }
    const meta = (loadedAccount as unknown as { meta?: Record<string, unknown> })
      .meta;
    const seenOnAccount = meta?.[HOW_TO_USE_SEEN_META_KEY] === true;

    if (seenOnAccount) return;

    const persist = () => {
      if (persisted.current) return;
      persisted.current = true;
      return accountUpdate({
        meta: { ...(meta ?? {}), [HOW_TO_USE_SEEN_META_KEY]: true },
      })
        .then(() => mutateAccount())
        .catch(() => {
          persisted.current = false;
        });
    };

    persist();
    setTooltip(true);
    trackHtuTooltipShown();
  }, [loadedAccount, mutateAccount, trackHtuTooltipShown]);

  useEffect(() => {
    if (!tooltip) return;
    const timer = window.setTimeout(
      () => setTooltip(false),
      HOW_TO_USE_TOOLTIP_MS,
    );
    return () => window.clearTimeout(timer);
  }, [tooltip]);

  const startTimer = useRef<number | null>(null);

  useEffect(
    () => () => {
      if (startTimer.current !== null) window.clearTimeout(startTimer.current);
    },
    [],
  );

  const open = useCallback(() => {
    setTooltip(false);
    trackHtuClicked();
    if (startTimer.current !== null) window.clearTimeout(startTimer.current);
    const scrolled = scrollToTop();
    if (!scrolled) {
      setPhase(0);
      return;
    }
    startTimer.current = window.setTimeout(() => {
      startTimer.current = null;
      scrollToTop();
      setPhase(0);
    }, EXPAND_SETTLE_MS);
  }, [trackHtuClicked]);

  const next = useCallback(() => {
    if (phase === null) return;
    trackHtuNextClicked(phase + 1);
    setPhase(phase + 1 < HOW_TO_USE_STEPS.length ? phase + 1 : null);
  }, [phase, trackHtuNextClicked]);

  const dismiss = useCallback(() => {
    if (phase === null) return;
    trackHtuDismissed(phase + 1);
    setPhase(null);
  }, [phase, trackHtuDismissed]);

  return (
    <>
      <styled.div position="relative" display="inline-flex">
        <styled.button
          type="button"
          onClick={open}
          aria-label="How to use community"
          display="inline-flex"
          alignItems="center"
          justifyContent="center"
          w="10"
          h="10"
          style={{ background: "transparent", border: "none", cursor: "pointer", padding: 0 }}
        >
          <InfoIcon />
        </styled.button>
        {tooltip && (
          <styled.div
            role="tooltip"
            onClick={open}
            style={{
              position: "absolute",
              top: "calc(100% - 4px)",
              left: -5,
              width: 170,
              zIndex: 20,
              cursor: "pointer",
              transformOrigin: "30px 0",
              animation: "htuPop 260ms cubic-bezier(.22,1,.36,1) both",
              filter:
                "drop-shadow(0 3px 3px rgba(0,0,0,0.12)) drop-shadow(0 11px 5.5px rgba(0,0,0,0.1)) drop-shadow(0 24px 7px rgba(0,0,0,0.06)) drop-shadow(0 43px 8.5px rgba(0,0,0,0.02))",
            }}
          >
            <styled.div style={{ paddingLeft: 16, height: 9, lineHeight: 0 }}>
              <svg width="19" height="9" viewBox="0 0 19 9" fill="none" aria-hidden>
                <path
                  d="M8.25821 0.491241L0 9.00004H19C16.704 6.54558 11.8453 1.63313 10.7357 0.491241C10.0994 -0.163594 8.89434 -0.163899 8.25821 0.491241Z"
                  fill="white"
                />
              </svg>
            </styled.div>
            <styled.div
              style={{
                padding: 12,
                borderRadius: 16,
                background: "#ffffff",
                color: "#404040",
                fontSize: 14,
                lineHeight: "20px",
                fontWeight: 700,
              }}
            >
              Learn to use community features.
            </styled.div>
            <style>{`@keyframes htuPop{0%{opacity:0;transform:scale(.6) translateY(-6px)}100%{opacity:1;transform:scale(1) translateY(0)}}`}</style>
          </styled.div>
        )}
      </styled.div>
      {phase !== null && (
        <Spotlight
          account={account}
          key={phase}
          step={HOW_TO_USE_STEPS[phase]!}
          index={phase}
          total={HOW_TO_USE_STEPS.length}
          onNext={next}
          onDismiss={dismiss}
        />
      )}
    </>
  );
}
