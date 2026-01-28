import type { Metadata, Viewport } from "next";
import { PropsWithChildren } from "react";

import { getColourAsHex } from "src/utils/colour";

import { inter, interDisplay } from "@/app/fonts";
import { getServerSession } from "@/auth/server-session";
import { SettingsContext } from "@/components/site/SettingsContext/SettingsContext";
import { serverEnvironment } from "@/config";
import { getSettings } from "@/lib/settings/settings-server";
import { getIconURL } from "@/utils/icon";

import "./global.css";

import { Providers } from "./providers";

export default async function RootLayout({ children }: PropsWithChildren) {
  const session = await getServerSession();
  const settings = await getSettings();

  const colorModeAttr =
    settings.color_mode === "system"
      ? undefined
      : settings.color_mode === "dark"
        ? "dark"
        : "light";

  return (
    <html
      lang="en"
      className={`${inter.variable} ${interDisplay.variable}`}
      data-color-mode={colorModeAttr}
    >
      <head>
        {/*
          NOTE: Because the browser side does not support dynamic environment
          variables (obviously, it's a browser script) we hack around Next.js'
          build-time variables by loading a dynamic script that sets the
          API_ADDRESS and WEB_ADDRESS. The script is served from /config.js
          which is a dynamic route that reads env vars at request time.

          This must be a blocking script (no async/defer) to ensure it executes
          before any React code that might try to read the config.
        */}
        <script src="/config.js" />

        {/*
            NOTE: This stylesheet is fully server-side rendered but it's not
            static because it uses data from the API to be generated. But we
            don't want this to require client-side render or CSS-in-JS.
        */}
        {/* eslint-disable-next-line @next/next/no-css-tags */}
        <link rel="stylesheet" href="/theme.css" />
      </head>

      <body>
        <SettingsContext initialSession={session} initialSettings={settings}>
          <Providers>{children}</Providers>
        </SettingsContext>
      </body>
    </html>
  );
}

export async function generateViewport(): Promise<Viewport> {
  const settings = await getSettings();

  const themeColour = getColourAsHex(settings.accent_colour);

  return {
    themeColor: themeColour,
    colorScheme: "only light",
  };
}

export async function generateMetadata(): Promise<Metadata> {
  const settings = await getSettings();

  const iconURL = getIconURL("512x512");

  const { WEB_ADDRESS } = serverEnvironment();
  const canonical = WEB_ADDRESS;

  // TODO: Add another settings field for this.
  const title = `${settings.title} | ${settings.description}`;

  return {
    manifest: "/manifest.json",
    metadataBase: new URL(canonical),
    title: title,
    description: settings.description,
    icons: {
      icon: iconURL,
      shortcut: iconURL,
      apple: iconURL,
    },
    appleWebApp: {
      capable: true,
      title: title,
      statusBarStyle: "black-translucent",
      startupImage: iconURL,
    },
  };
}
