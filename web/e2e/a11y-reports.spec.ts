import { expect, test } from "@playwright/test";

import { expectAccessible } from "./a11y";

const TOKEN = "e2e-a11y-token";

const reportRoutes = [
  { path: "/reports", label: "reports index" },
  { path: "/reports/trial-balance", label: "trial balance" },
  { path: "/reports/aging", label: "aging report" },
  { path: "/reports/profit-and-loss", label: "profit & loss" },
  { path: "/reports/balance-sheet", label: "balance sheet" },
  { path: "/reports/cash-flow", label: "cash flow" },
  { path: "/reports/inventory-valuation", label: "inventory valuation" },
] as const;

test.describe("reports a11y", () => {
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

  for (const { path, label } of reportRoutes) {
    test(`${label} has no a11y violations`, async ({ page }) => {
      await page.goto(`${path}`);
      await expect(page.getByRole("main").first()).toBeVisible();
      await expectAccessible(page, { ruleTags: ["wcag2a", "wcag2aa"] });
    });
  }
});
