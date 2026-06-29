import {
  Sun,
  PlaneTakeoff,
  CalendarDays,
  UserRoundCog,
  Salad,
  HandMetal,
} from "lucide-react";

export type Gender = "M" | "F";

/** Any lucide-react icon satisfies this shape. */
export type TopicIcon = React.ComponentType<{
  size?: number;
  color?: string;
  strokeWidth?: number;
  absoluteStrokeWidth?: boolean;
}>;

export interface TipTopic {
  id: string;
  title: string;
  Icon: TopicIcon;
  prompt: string;
  placeholder: string;
  suggestions: string[];
}

export const TOPICS: TipTopic[] = [
  {
    id: "busy-morning",
    title: "Busy morning routine tips",
    Icon: Sun,
    prompt: "Tell us how you stick to your routine on rushed mornings",
    placeholder: "e.g. me alarm set karti hu",
    suggestions: [
      "me alarm set karti hu",
      "brush ke saath hi le leti hu",
      "raat ko sab nikaal ke rakhti hu",
      "2-min wala shortcut",
    ],
  },
  {
    id: "travel-routine",
    title: "Travel routine tips",
    Icon: PlaneTakeoff,
    prompt: "Tell us how you keep your routine going while travelling",
    placeholder: "e.g. ek week ki dose pack karti hu",
    suggestions: [
      "ek week ki dose pack karti hu",
      "sab ek pouch mein pack karti hu",
      "hotel mein bhi miss nahi karta",
      "hand-bag mein hi rakhta hu",
    ],
  },
  {
    id: "first-month",
    title: "First month tips",
    Icon: CalendarDays,
    prompt: "Tell us what got you through your first month",
    placeholder: "e.g. daily routine fix kiya",
    suggestions: [
      "pehle mahine patience rakha aur routine follow kiya",
      "coach se regular baat ki",
      "shedding se ghabrayi nahi",
      "daily routine fix kiya",
    ],
  },
  {
    id: "small-habit",
    title: "One small habit, big difference",
    Icon: UserRoundCog,
    prompt: "Tell us one small habit that made a big difference",
    placeholder: "e.g. phone reminder laga diya",
    suggestions: [
      "daily fixed time set kiya",
      "phone reminder laga diya",
      "kit sink ke paas rakhi",
      "paani zyada peena shuru kiya",
    ],
  },
  {
    id: "diet-change",
    title: "One diet change",
    Icon: Salad,
    prompt: "Tell us one diet change that helped you",
    placeholder: "e.g. protein badha diya",
    suggestions: [
      "protein badha diya",
      "bahar ka khana kam kar diya",
      "badaam khana shuru kiya",
      "hari sabjiya add ki khane mai",
    ],
  },
  {
    id: "staying-motivated",
    title: "Staying motivated early on",
    Icon: HandMetal,
    prompt: "Tell us how you stayed motivated before results showed",
    placeholder: "e.g. progress photos lete raha",
    suggestions: [
      "progress photos lete raha",
      "results aate hai, yaad rakha",
      "chhote changes notice kiye",
      "quit ka mann hua to doosro ke results se motivate hua",
    ],
  },
];

export const getTopic = (id: string | null): TipTopic | undefined =>
  TOPICS.find((t) => t.id === id);

export interface TipsTheme {
  pageGradient: string;
  selectedBorder: string;
  selectedFill: string;
  radio: string;
  imageBtnBg: string;
  imageBtnFg: string;
  successColor: string;
}

export const getTheme = (gender: Gender): TipsTheme =>
  gender === "F"
    ? {
        pageGradient: "linear-gradient(180deg, #F7E6D4 0px, #FFFFFF 254px)",
        selectedBorder: "#A54337",
        selectedFill: "#FBEEDD",
        radio: "#A54337",
        imageBtnBg: "#EEEAF4",
        imageBtnFg: "#998AB3",
        successColor: "#3BA773",
      }
    : {
        pageGradient: "linear-gradient(180deg, #D6EAE0 0px, #FFFFFF 254px)",
        selectedBorder: "#2D8A5D",
        selectedFill: "#EAF5F0",
        radio: "#2D8A5D",
        imageBtnBg: "#DAEBF1",
        imageBtnFg: "#519FBD",
        successColor: "#3BA773",
      };

export const INK = "#2C2C2A";

export const CARD_BORDER = "#F0F0F0";

export const FONT = "var(--font-nunito-sans), 'Nunito Sans', sans-serif";

// Traya grey scale used across the tip flow. Kept here as plain values rather
// than a parallel CSS-variable token layer.
export const GREY = {
  0: "#ffffff",
  100: "#f0f0f0",
  200: "#dedede",
  300: "#c9c9c9",
  400: "#b5b5b5",
  500: "#999999",
  600: "#787878",
  650: "#5c5c5c",
  700: "#404040",
  800: "#303030",
} as const;
