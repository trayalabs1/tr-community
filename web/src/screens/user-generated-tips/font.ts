import localFont from "next/font/local";

// Nunito Sans (variable, latin) — self-hosted to match the Figma type spec.
export const nunitoSans = localFont({
  src: [{ path: "fonts/NunitoSans-latin.woff2", weight: "400 700" }],
  preload: true,
  display: "swap",
  variable: "--font-nunito-sans",
});
