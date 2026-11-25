
export type HowToUseStep = {
  target: string;
  title: string;
  items: { icon: "search" | "bell" | "profile" | "send" | "help" | "filter" | "posts" | "bookmark" | "like" | "comment" | "share"; label: string }[];
  showPointer?: boolean;
  padding?: { x: number; y: number };
};

export const HOW_TO_USE_STEPS: HowToUseStep[] = [
  {
    target: "topbar",
    title: "Explore from the top bar",
    showPointer: true,
    padding: { x: -8, y: -3 },
    items: [
      { icon: "search", label: "Search posts" },
      { icon: "bell", label: "Get updates" },
      { icon: "profile", label: "View or edit your profile" },
    ],
  },
  {
    target: "composer",
    title: "Quick actions",
    items: [
      { icon: "help", label: "Ask a question" },
      { icon: "send", label: "Tap send to share" },
      { icon: "help", label: "What to post? Get ideas here" },
    ],
  },
  {
    target: "filters",
    title: "Find posts faster",
    padding: { x: -8, y: -6 },
    items: [
      { icon: "filter", label: "Search by filters" },
      { icon: "posts", label: "See your posts" },
      { icon: "bookmark", label: "Saved posts" },
    ],
  },
  {
    target: "post",
    title: "Engage with posts",
    items: [
      { icon: "like", label: "Like a post" },
      { icon: "comment", label: "Comment" },
      { icon: "bookmark", label: "Save post" },
      { icon: "share", label: "Share" },
    ],
  },
];

export const HOW_TO_USE_TOOLTIP_MS = 5000;
export const HOW_TO_USE_SEEN_META_KEY = "community_htu_seen";

export const HOW_TO_USE_SHOW_META_KEY = "community_htu_show";
