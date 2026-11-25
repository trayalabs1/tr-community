import { z } from "zod";

const ConfigSchema = z.object({
  API_ADDRESS: z.string().url("API_ADDRESS must be a valid URL"),
  WEB_ADDRESS: z.string().url("WEB_ADDRESS must be a valid URL"),
  source: z.union([z.literal("server"), z.literal("script")]),
});
export type Config = z.infer<typeof ConfigSchema>;

export function serverEnvironment(): Config {
  const apiAddress = process.env["NEXT_PUBLIC_API_ADDRESS"];
  const webAddress = process.env["NEXT_PUBLIC_WEB_ADDRESS"];

  const parsed = ConfigSchema.safeParse({
    API_ADDRESS: apiAddress,
    WEB_ADDRESS: webAddress,
    source: "server" as const,
  });

  if (!parsed.success) {
    throw new Error(
      `Invalid configuration: Missing or invalid required environment variables.\n` +
      `Required: NEXT_PUBLIC_API_ADDRESS, NEXT_PUBLIC_WEB_ADDRESS\n` +
      `Errors: ${parsed.error.message}`
    );
  }

  return parsed.data;
}

function isomorphicEnvironment(): Config {
  if (typeof window !== "undefined") {
    const parsed = ConfigSchema.safeParse((window as any).__storyden__);

    if (!parsed.success) {
      throw new Error(
        `Invalid config from window.__storyden__. This indicates the root layout <script> tag failed to inject configuration. ` +
        `Error: ${parsed.error.message}`
      );
    }

    return parsed.data;
  } else {
    return serverEnvironment();
  }
}

let cachedEnv: Config | null = null;

function getEnv(): Config {
  if (cachedEnv === null) {
    cachedEnv = isomorphicEnvironment();
  }
  return cachedEnv;
}

export function getWEBAddress() {
  return getEnv().WEB_ADDRESS;
}

export function getAPIAddress() {
  const env = getEnv();
  const isClient = typeof window !== "undefined";
  const result = isClient
    ? env.API_ADDRESS
    : global.process.env["SSR_API_ADDRESS"] ?? env.API_ADDRESS;

  return result;
}

// Deprecated: Use getAPIAddress() and getWEBAddress() instead
// These are kept for backward compatibility using lazy getters
export const API_ADDRESS = getAPIAddress();
export const WEB_ADDRESS = getWEBAddress();
