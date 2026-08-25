// Mirrors the v4 sentiment taxonomy in app/services/sentiment/scorer/scorer.go
// (the Category type/consts), excluding "NA" — posts without a classified
// topic aren't a filterable topic, they just show up under "All".
export const PRIMARY_TOPICS: { label: string; value: string }[] = [
  { label: "Results & Progress", value: "RESULTS & PROGRESS" },
  { label: "Tips & Experiences", value: "TIPS & EXPERIENCES" },
  { label: "Hairfall Concerns", value: "HAIRFALL CONCERNS" },
  { label: "How To Use", value: "HOW TO USE" },
  { label: "Dandruff & Scalp", value: "DANDRUFF & SCALP" },
  { label: "Hair Regrowth", value: "HAIR REGROWTH" },
  { label: "Products & Treatment", value: "PRODUCTS & TREATMENT" },
  { label: "Side Effects", value: "SIDE EFFECTS" },
  { label: "Diet & Lifestyle", value: "DIET & LIFESTYLE" },
  { label: "Challenges", value: "CHALLENGES" },
];
