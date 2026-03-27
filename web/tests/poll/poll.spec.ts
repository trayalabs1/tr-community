import { BrowserContext, Page, expect, test } from "@playwright/test";

import {
  createAccountWithRole,
  login,
  logout,
  registerUser,
} from "../access_key_admin_assignment";

const PASSWORD = "TestPassword123!";

async function dismissOnboarding(page: Page) {
  const skipButton = page.getByRole("button", { name: "Skip" });
  if (await skipButton.isVisible({ timeout: 1000 }).catch(() => false)) {
    await skipButton.click();
  }
}

async function loginViaPage(page: Page, username: string) {
  await page.goto("/login");
  await page.getByRole("textbox", { name: "username" }).fill(username);
  await page.getByRole("textbox", { name: "password" }).fill(PASSWORD);
  await page.getByRole("button", { name: "Login" }).click();
  await expect(page).toHaveURL("/", { timeout: 10000 });
  await expect(
    page.getByRole("button", { name: "Account menu" }),
  ).toBeVisible();
  await dismissOnboarding(page);
}

async function openComposeModal(page: Page) {
  await page.getByRole("link", { name: "Post" }).click();
  await expect(page).toHaveURL("/new", { timeout: 5000 });
}

async function selectPollsCategory(page: Page) {
  const categorySelect = page.getByRole("combobox").first();
  await categorySelect.click();
  await page.getByRole("option", { name: /polls?/i }).click();
}

async function fillPollComposer(
  page: Page,
  question: string,
  options: string[],
) {
  await page
    .getByPlaceholder("Ask your question...")
    .fill(question);

  for (let i = 0; i < options.length; i++) {
    await page.getByPlaceholder(`Option ${i + 1}`).fill(options[i]!);
    if (i < options.length - 1) {
      await page.getByRole("button", { name: "+ Add option" }).click();
    }
  }
}

test.describe("Polls", () => {
  test("admin creates a poll, member votes on it", async ({ browser }) => {
    const ts = Date.now();
    const adminHandle = `poll-admin-${ts}`;
    const memberHandle = `poll-voter-${ts}`;
    const pollQuestion = `Which fruit do you prefer? ${ts}`;

    const adminCtx = await browser.newContext();
    const memberCtx = await browser.newContext();

    await createAccountWithRole(adminCtx, adminHandle, PASSWORD, "admin");
    await createAccountWithRole(memberCtx, memberHandle, PASSWORD, "member");

    const adminPage = await adminCtx.newPage();
    await loginViaPage(adminPage, adminHandle);

    await openComposeModal(adminPage);
    await selectPollsCategory(adminPage);

    await expect(
      adminPage.getByPlaceholder("Ask your question..."),
    ).toBeVisible({ timeout: 5000 });

    await fillPollComposer(adminPage, pollQuestion, ["Apple", "Banana"]);

    await adminPage.getByRole("button", { name: "Post" }).click();
    await expect(adminPage).toHaveURL("/", { timeout: 10000 });

    const pollCard = adminPage.locator("div").filter({ hasText: pollQuestion }).first();
    await expect(pollCard).toBeVisible({ timeout: 10000 });

    await logout(adminPage);

    const memberPage = await memberCtx.newPage();
    await loginViaPage(memberPage, memberHandle);

    const memberPollCard = memberPage
      .locator("div")
      .filter({ hasText: pollQuestion })
      .first();
    await expect(memberPollCard).toBeVisible({ timeout: 10000 });

    const appleButton = memberPollCard.getByRole("button", { name: "Apple" });
    const bananaButton = memberPollCard.getByRole("button", { name: "Banana" });
    await expect(appleButton).toBeVisible();
    await expect(bananaButton).toBeVisible();

    await appleButton.click();

    await expect(
      memberPollCard.getByText("Apple"),
    ).toBeVisible({ timeout: 5000 });
    await expect(memberPollCard.getByText("1")).toBeVisible({ timeout: 5000 });
    await expect(memberPollCard.getByText(/\d+ vote/)).toBeVisible({
      timeout: 5000,
    });

    await expect(
      memberPollCard.getByRole("button", { name: "Apple" }),
    ).not.toBeVisible();

    await adminCtx.close();
    await memberCtx.close();
  });
});
