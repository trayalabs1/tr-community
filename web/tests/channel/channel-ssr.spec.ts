import { Page, expect, test } from "@playwright/test";

const PASSWORD = "TestPassword123!";

async function dismissOnboarding(page: Page) {
  const skipButton = page.getByRole("button", { name: "Skip" });
  if (await skipButton.isVisible({ timeout: 1000 }).catch(() => false)) {
    await skipButton.click();
  }
}

async function registerUser(page: Page, username: string) {
  await page.goto("/register");
  await page.getByRole("textbox", { name: "username" }).fill(username);
  await page.getByRole("textbox", { name: "password" }).fill(PASSWORD);
  await page.getByRole("button", { name: "Register" }).click();
  await expect(page).toHaveURL("/", { timeout: 10000 });
  await expect(
    page.getByRole("button", { name: "Account menu" }),
  ).toBeVisible();
  await dismissOnboarding(page);
}

async function openFirstChannel(page: Page): Promise<string> {
  await page.goto("/channels");
  await dismissOnboarding(page);

  const card = page.locator("[aria-label$='channel']").first();
  await expect(card).toBeVisible({ timeout: 10000 });
  await card.click();

  await expect(page).toHaveURL(/\/channels\/[^/]+$/, { timeout: 10000 });
  return page.url();
}

async function createThreadInChannel(
  page: Page,
  title: string,
  body: string,
) {
  await page.getByRole("link", { name: "Post" }).first().click();
  await expect(page).toHaveURL(/\/new/, { timeout: 5000 });

  await page.locator("#title-input").fill(title);

  const editor = page.locator(".ProseMirror").first();
  await editor.click();
  await editor.fill(body);

  await page.getByRole("button", { name: "Post" }).click();
  await expect(page).toHaveURL(/\/t\//, { timeout: 10000 });
  await dismissOnboarding(page);
}

test.describe("Channel SSR", () => {
  test("first page of threads is server-rendered into the HTML", async ({
    page,
  }) => {
    const title = `SSR Thread ${Date.now()}`;

    await registerUser(page, "channel-ssr-01");
    const channelUrl = await openFirstChannel(page);
    await createThreadInChannel(page, title, "This thread must appear in SSR.");

    // Fetch the raw server HTML for the channel page, carrying the session
    // cookies but WITHOUT executing any client JS. If threads are prefetched
    // server-side and seeded into SWR (fix #3), the title is present in the
    // initial document. Without the fix the document only contains the
    // "Loading threads..." placeholder.
    const response = await page.request.get(channelUrl);
    expect(response.ok()).toBeTruthy();
    const html = await response.text();

    expect(html).toContain(title);
    expect(html).not.toContain("Loading threads...");
  });

  test("thread is visible after navigating to the channel", async ({
    page,
  }) => {
    const title = `Visible Thread ${Date.now()}`;

    await registerUser(page, "channel-ssr-02");
    const channelUrl = await openFirstChannel(page);
    await createThreadInChannel(page, title, "This thread must render.");

    await page.goto(channelUrl);
    await dismissOnboarding(page);

    await expect(
      page.locator("main").getByText(title),
    ).toBeVisible({ timeout: 10000 });
  });
});
