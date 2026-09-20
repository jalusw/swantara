import { expect, test } from "@playwright/test";

import { expectAccessible } from "./a11y";

// Sidebar: `protectedPaths` + `protectedPrefixes` in lib/middleware/auth.ts
// require an `access_token` cookie to avoid redirect to /login.
// We set a dummy token so the shell renders (data fetches degrade via allSettled)
// and we can still axe-scan the layout/shell itself.

const TOKEN = "e2e-a11y-token";

const protectedRoutes = [
  { path: "/onboarding", label: "onboarding" },
  { path: "/dashboard", label: "org dashboard" },
  { path: "/products", label: "org products" },
  { path: "/settings", label: "org settings" },
] as const;

test.describe("protected", () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([
      {
        name: "access_token",
        value: TOKEN,
        domain: "localhost",
        path: "/",
      },
    ]);
  });

  for (const { path, label } of protectedRoutes) {
    test(`${label} has no a11y violations`, async ({ page }) => {
      await page.goto(`${path}`);

      // Org shell and onboarding both render a <main> even when backend is down
      await expect(page.getByRole("main").first()).toBeVisible();

      await expectAccessible(page, { ruleTags: ["wcag2a", "wcag2aa"] });
    });
  }
});
