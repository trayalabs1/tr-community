import { NextResponse } from "next/server";

import { serverEnvironment } from "@/config";

export const dynamic = "force-dynamic";
export const revalidate = 0;

export async function GET() {
  const { API_ADDRESS, WEB_ADDRESS, HOW_TO_USE_NEW_CUSTOMER_CUTOFF_DATE } =
    serverEnvironment();

  console.log(
    "[HowToUse] config.js served, HOW_TO_USE_NEW_CUSTOMER_CUTOFF_DATE =",
    HOW_TO_USE_NEW_CUSTOMER_CUTOFF_DATE ?? "(unset)",
  );

  const configJson = JSON.stringify({
    API_ADDRESS,
    WEB_ADDRESS,
    HOW_TO_USE_NEW_CUSTOMER_CUTOFF_DATE,
    source: "script",
  });

  const script = `window.__storyden__ = ${configJson};
console.log('[Config] Set window.__storyden__:', window.__storyden__);`;

  return new NextResponse(script, {
    headers: {
      "Content-Type": "application/javascript",
      "Cache-Control": "no-store, must-revalidate",
    },
  });
}
