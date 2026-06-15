import {
  HelpCircle,
  type LucideIcon,
  Package,
  Sparkles,
  TimerReset,
  Users,
} from "lucide-react";

import { Metadata } from "@/api/openapi-schema";

export type PromptItem = {
  icon: string;
  text: string;
  tag?: string;
};

export const PROMPT_ICONS: Record<string, LucideIcon> = {
  kit: Sparkles,
  products: Package,
  clock: TimerReset,
  help: HelpCircle,
  people: Users,
};

export const PROMPT_ICON_KEYS = Object.keys(PROMPT_ICONS);

export const DEFAULT_PROMPT_ICON = "kit";

export function resolvePromptIcon(icon: string): LucideIcon {
  return PROMPT_ICONS[icon] ?? PROMPT_ICONS[DEFAULT_PROMPT_ICON]!;
}

export function parsePromptNudges(meta?: Metadata): PromptItem[] {
  const raw = meta?.["prompt_nudges"];
  if (!Array.isArray(raw)) {
    return [];
  }

  return raw.flatMap((item): PromptItem[] => {
    if (typeof item !== "object" || item === null) {
      return [];
    }

    const { icon, text, tag } = item as Record<string, unknown>;
    if (typeof text !== "string" || text.trim() === "") {
      return [];
    }

    return [
      {
        icon: typeof icon === "string" ? icon : DEFAULT_PROMPT_ICON,
        text,
        ...(typeof tag === "string" && tag.trim() !== "" ? { tag } : {}),
      },
    ];
  });
}
