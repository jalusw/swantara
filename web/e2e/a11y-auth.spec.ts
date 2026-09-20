import { expect, test } from "@playwright/test";

import { expectAccessible } from "./a11y";

const authRoutes = [
  { path: "/login", label: "login" },
  { path: "/register", label: "register" },
  { path: "/forgot-password", label: "forgot-password" },
  { path: "/reset-password", label: "reset-password" },
] as const;

test.describe("auth", () => {
  for (const { path, label } of authRoutes) {
    test(`${label} has no a11y violations`, async ({ page }) => {
      await page.goto(`${path}`);

      // Auth pages render inside AuthShell's <main>
      await expect(page.getByRole("main")).toBeVisible();

      await expectAccessible(page, { ruleTags: ["wcag2a", "wcag2aa"] });
    });
  }
});
