import { expect, test } from "@playwright/test";

import { expectAccessible } from "./a11y";

test("root redirects to login", async ({ page }) => {
  await page.goto("/");

  await page.waitForURL(/\/login/);
  await expect(page.getByRole("main")).toBeVisible();
});

test("login page has no accessibility violations", async ({ page }) => {
  await page.goto("/login");

  await expect(page.getByRole("main")).toBeVisible();

  await expectAccessible(page, { ruleTags: ["wcag2a", "wcag2aa"] });
});
