import { APIRequestContext, Page, expect, request, test } from "@playwright/test";

import { createAccountWithRole } from "../access_key_admin_assignment";

const PASSWORD = "TestPassword123!";

function apiUrl(): string {
  return process.env["PUBLIC_API_ADDRESS"] || "http://localhost:8001";
}

async function dismissOnboarding(page: Page) {
  const skip = page.getByRole("button", { name: "Skip" });
  if (await skip.isVisible({ timeout: 1000 }).catch(() => false)) {
    await skip.click();
  }
}

// An authenticated API context for driving setup as a given account.
async function apiSession(username: string): Promise<APIRequestContext> {
  const ctx = await request.newContext({ baseURL: apiUrl() });
  const resp = await ctx.post("/api/auth/password/signin", {
    data: { identifier: username, token: PASSWORD },
  });
  if (!resp.ok()) {
    throw new Error(`signin ${username}: ${resp.status()} ${await resp.text()}`);
  }
  return ctx;
}

async function createChannel(
  ctx: APIRequestContext,
  name: string,
  slug: string,
): Promise<string> {
  const resp = await ctx.post("/api/channels", {
    data: { name, slug, description: name },
  });
  expect(resp.ok(), `create channel ${slug}`).toBeTruthy();
  return (await resp.json()).id;
}

async function selfAccountID(ctx: APIRequestContext): Promise<string> {
  const resp = await ctx.get("/api/accounts");
  expect(resp.ok(), "self account").toBeTruthy();
  return (await resp.json()).id;
}

test.describe("Interchannel thread sharing", () => {
  test("cross-cohort reply shows the origin cohort pill", async ({
    page,
  }) => {
    test.setTimeout(90000);

    const ts = Date.now();
    const adminHandle = `share_admin_${ts}`;
    const memberHandle = `share_member_${ts}`;

    // Create both accounts via the e2e access-key helper (admin gets the role).
    const setupCtx = await page.context().browser()!.newContext();
    await createAccountWithRole(setupCtx, adminHandle, PASSWORD, "admin");
    await createAccountWithRole(setupCtx, memberHandle, PASSWORD, "member");
    await setupCtx.close();

    const admin = await apiSession(adminHandle);
    const member = await apiSession(memberHandle);
    const memberAccountID = await selfAccountID(member);

    // Two cohorts, named by month to mirror the product's cohort naming.
    const sourceChannel = await createChannel(
      admin,
      `Month 8 Warriors ${ts}`,
      `m8-${ts}`,
    );
    const destChannel = await createChannel(
      admin,
      `Month 2 Community ${ts}`,
      `m2-${ts}`,
    );

    // The member belongs to the destination cohort.
    const addResp = await admin.post(
      `/api/channels/${destChannel}/members`,
      { data: { account_id: memberAccountID, role: "member" } },
    );
    expect(addResp.ok(), "add member to destination").toBeTruthy();

    // A published thread lives in the source cohort.
    const threadResp = await admin.post(
      `/api/channels/${sourceChannel}/threads`,
      {
        data: {
          title: `Hair regrowth finally visible ${ts}`,
          body: "<p>8 months of consistency, before and after in comments.</p>",
          visibility: "published",
        },
      },
    );
    expect(threadResp.ok(), "create source thread").toBeTruthy();
    const threadSlug: string = (await threadResp.json()).slug;

    // Admin shares it into the destination cohort.
    const shareResp = await admin.post(`/api/threads/${threadSlug}/shares`, {
      data: { channels: [destChannel], subtitle: "Sharing for cohort 2" },
    });
    expect(shareResp.ok(), "share into destination").toBeTruthy();

    // Authenticate the browser as the member by signing in through the page's
    // own request context, so the Set-Cookie lands natively in the browser jar.
    const uiSignin = await page.request.post(
      `${apiUrl()}/api/auth/password/signin`,
      { data: { identifier: memberHandle, token: PASSWORD } },
    );
    expect(uiSignin.ok(), "browser signin as member").toBeTruthy();

    // --- UI: member replies from the DESTINATION channel route ---
    const replyBody = `Reply from cohort two ${ts}`;
    await page.goto(`/channels/${destChannel}/threads/${threadSlug}`);
    await dismissOnboarding(page);

    const replyForm = page
      .locator("form")
      .filter({ has: page.getByRole("button", { name: "Post" }) })
      .last();
    const replyEditor = replyForm.locator(
      ".ProseMirror[contenteditable='true']",
    );
    await replyEditor.click();
    await replyEditor.fill(replyBody);
    await replyForm.getByRole("button", { name: "Post" }).click();

    const reply = page.locator("li").filter({ hasText: replyBody });
    await expect(reply).toBeVisible({ timeout: 10000 });

    // The reply's byline must carry the origin cohort pill (the destination
    // channel name), proving the reply recorded its origin cohort.
    await expect(reply.getByText(/Month 2 Community/i).first()).toBeVisible({
      timeout: 10000,
    });

    await admin.dispose();
    await member.dispose();
  });
});
