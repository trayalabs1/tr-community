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
  await dismissOnboarding(page);
}

test.describe("search screen", () => {
  test("shows the empty state before a query is entered", async ({ page }) => {
    await registerUser(page, `search-empty-${Date.now()}`);
    await page.goto("/search");

    await expect(
      page.getByText("Search post across all channels"),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "Clear" })).toBeHidden();
  });

  test("reveals the Clear action once a query is present", async ({ page }) => {
    await registerUser(page, `search-clear-${Date.now()}`);
    await page.goto("/search?q=traya");

    await expect(page.getByRole("button", { name: "Clear" })).toBeVisible();
    await expect(
      page.getByText("Search post across all channels"),
    ).toBeHidden();
  });

  test("clearing the query restores the empty state", async ({ page }) => {
    await registerUser(page, `search-restore-${Date.now()}`);
    await page.goto("/search?q=traya");

    await page.getByRole("button", { name: "Clear" }).click();

    await expect(
      page.getByText("Search post across all channels"),
    ).toBeVisible();
  });

  test("submitting a query updates the URL", async ({ page }) => {
    await registerUser(page, `search-submit-${Date.now()}`);
    await page.goto("/search");

    await page.getByPlaceholder("Search posts...").fill("traya");
    await page.getByPlaceholder("Search posts...").press("Enter");

    await expect(page).toHaveURL(/q=traya/, { timeout: 10000 });
  });

  test("renders at mobile width without horizontal overflow", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 390, height: 800 });
    await registerUser(page, `search-mobile-${Date.now()}`);
    await page.goto("/search");

    await expect(
      page.getByText("Search post across all channels"),
    ).toBeVisible();

    const overflows = await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth + 1,
    );
    expect(overflows).toBe(false);
  });
});
