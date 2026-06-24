import { expect, test } from "@playwright/test";

import { createAccountWithRole, login } from "../access_key_admin_assignment";

const PASSWORD = "TestPassword123!";

test.describe("Bulk Actions queue tab", () => {
  test("gates the panel until a post type is selected", async ({ browser }) => {
    const timestamp = Date.now();
    const adminHandle = `bulk_admin_${timestamp}`;

    const adminCtx = await browser.newContext();
    await createAccountWithRole(adminCtx, adminHandle, PASSWORD, "admin");

    const adminPage = await adminCtx.newPage();
    await login(adminPage, adminHandle, PASSWORD);
    await adminPage.goto("/queue");

    await adminPage
      .getByRole("button", { name: "Bulk Actions", exact: true })
      .click();

    // Before selecting a post type, the gated empty state shows and the
    // bulk action buttons are not available.
    await expect(
      adminPage.getByText("Select a post type to load threads"),
    ).toBeVisible();
    await expect(
      adminPage.getByRole("button", { name: "Bulk Like" }),
    ).toHaveCount(0);
    await expect(
      adminPage.getByRole("button", { name: "Bulk Reply" }),
    ).toHaveCount(0);

    // Selecting "Streak Posts" reveals the bulk action buttons.
    await adminPage.getByRole("button", { name: "Streak Posts" }).click();

    await expect(
      adminPage.getByRole("button", { name: "Bulk Like" }),
    ).toBeVisible({ timeout: 10000 });
    await expect(
      adminPage.getByRole("button", { name: "Bulk Reply" }),
    ).toBeVisible();

    await adminCtx.close();
  });

  test("Bulk Reply asks for confirmation with a count", async ({ browser }) => {
    const timestamp = Date.now();
    const adminHandle = `bulk_confirm_admin_${timestamp}`;

    const adminCtx = await browser.newContext();
    await createAccountWithRole(adminCtx, adminHandle, PASSWORD, "admin");

    const adminPage = await adminCtx.newPage();
    await login(adminPage, adminHandle, PASSWORD);
    await adminPage.goto("/queue");

    await adminPage
      .getByRole("button", { name: "Bulk Actions", exact: true })
      .click();
    await adminPage.getByRole("button", { name: "Streak Posts" }).click();

    const bulkReply = adminPage.getByRole("button", { name: "Bulk Reply" });
    await expect(bulkReply).toBeVisible({ timeout: 10000 });

    // With no matching streak posts the action is disabled; with matches it
    // opens a confirmation prompt. Either way it must never navigate away.
    const isDisabled = await bulkReply.isDisabled();
    if (!isDisabled) {
      await bulkReply.click();
      await expect(adminPage.getByText(/Reply to \d+ thread/)).toBeVisible();
      await expect(
        adminPage.getByRole("button", { name: "Confirm" }),
      ).toBeVisible();
      await expect(
        adminPage.getByRole("button", { name: "Cancel" }),
      ).toBeVisible();
    }

    await expect(adminPage).toHaveURL(/\/queue$/);

    await adminCtx.close();
  });
});
