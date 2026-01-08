import { z } from "zod";

export const DEFAULT_API_ADDRESS = "http://localhost:8000";
export const DEFAULT_WEB_ADDRESS = "http://localhost:3000";

export const ConfigSchema = z.object({
  API_ADDRESS: z.string(),
  WEB_ADDRESS: z.string(),
  source: z.union([z.literal("server"), z.literal("script")]),
});
export type Config = z.infer<typeof ConfigSchema>;

export function serverEnvironment() {
  const apiAddress =
    global.process.env["NEXT_PUBLIC_API_ADDRESS"] ??
    global.process.env["PUBLIC_API_ADDRESS"] ??
    DEFAULT_API_ADDRESS;

  const webAddress =
    global.process.env["NEXT_PUBLIC_WEB_ADDRESS"] ??
    global.process.env["PUBLIC_WEB_ADDRESS"] ??
    DEFAULT_WEB_ADDRESS;

  console.log("[Config] Environment variables:", {
    PUBLIC_API_ADDRESS: global.process.env["PUBLIC_API_ADDRESS"],
    PUBLIC_WEB_ADDRESS: global.process.env["PUBLIC_WEB_ADDRESS"],
    SSR_API_ADDRESS: global.process.env["SSR_API_ADDRESS"],
    resolved_API_ADDRESS: apiAddress,
    resolved_WEB_ADDRESS: webAddress,
  });

  return {
    API_ADDRESS: apiAddress,
    WEB_ADDRESS: webAddress,
    source: "server" as const,
  };
}

function isomorphicEnvironment(): Config {
  if (typeof window !== "undefined") {
    const parsed = ConfigSchema.safeParse((window as any).__storyden__);

    // Don't bail out if the config is invalid, just log it and use the default.
    // This is a rare occurrance but it does happen when Next.js renders either
    // not-found.tsx or error.tsx as it does not load the <script> tag for some
    // weird unknown reason. This seems like a bug in Next.js and not Storyden.
    // Either way, the config isn't necessary on either of those pages as they
    // don't make client side API calls, they just render a basic error page.
    if (!parsed.success) {
      console.error(
        `Invalid config loaded from \`window.__storyden__\`, this indicates a problem running the root layout <script> tag on the server render which should inject the API_ADDRESS and WEB_ADDRESS environment variables.

A default configuration will be used, however this configuration will most likely not work correctly in most production environments.

If you see this, please open an issue at https://github.com/Southclaws/storyden/issues/new
`,
        parsed.error.issues,
      );

      return {
        API_ADDRESS: DEFAULT_API_ADDRESS,
        WEB_ADDRESS: DEFAULT_WEB_ADDRESS,
        source: "server",
      };
    }

    const config = parsed.data;
    console.log("loaded window config", config);
    return config;
  } else {
    const config = serverEnvironment();
    return config;
  }
}

const env = isomorphicEnvironment();

export const API_ADDRESS = env.API_ADDRESS;

export const WEB_ADDRESS = env.WEB_ADDRESS;

export function getAPIAddress() {
  const isClient = typeof window !== "undefined";
  const result = isClient
    ? env.API_ADDRESS
    : global.process.env["SSR_API_ADDRESS"] ?? env.API_ADDRESS;

  console.log("[getAPIAddress]", {
    isClient,
    result,
    envAPIAddress: env.API_ADDRESS,
    ssrAPIAddress: global.process.env["SSR_API_ADDRESS"],
    windowConfig: isClient ? (window as any).__storyden__ : "N/A (server)",
  });

  return result;
}
