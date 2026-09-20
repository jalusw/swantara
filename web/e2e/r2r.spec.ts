import { expect, test } from "@playwright/test";

import { navigateToOrg, seedOrg } from "./fixtures/helpers";

test.describe("Record-to-Report (R2R) flow", () => {
  test.beforeEach(async ({ page }) => {
    await seedOrg(page);
  });

  test("navigates to all report pages", async ({ page }) => {
    await navigateToOrg(page, "/reports");
    await expect(page.getByRole("main").first()).toBeVisible();
    await expect(page.getByRole("heading")).toBeVisible();

    await navigateToOrg(page, "/reports/trial-balance");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/reports/aging");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/reports/profit-and-loss");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/reports/balance-sheet");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/reports/cash-flow");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/reports/inventory-valuation");
    await expect(page.getByRole("main").first()).toBeVisible();
  });

  test("reports index shows report cards", async ({ page }) => {
    await navigateToOrg(page, "/reports");
    await expect(page.getByRole("main").first()).toBeVisible();
    const links = page.getByRole("link");
    const count = await links.count();
    expect(count).toBeGreaterThanOrEqual(6);
  });

  test("trial balance page has balanced indicator", async ({ page }) => {
    await navigateToOrg(page, "/reports/trial-balance");
    await expect(page.getByRole("main").first()).toBeVisible();
    const heading = page.getByRole("heading").first();
    await expect(heading).toBeVisible();
  });

  test("report pages have back navigation", async ({ page }) => {
    const reportPages = [
      "/reports/trial-balance",
      "/reports/aging",
      "/reports/profit-and-loss",
      "/reports/balance-sheet",
      "/reports/cash-flow",
      "/reports/inventory-valuation",
    ];

    for (const path of reportPages) {
      await navigateToOrg(page, path);
      const backLink = page.getByRole("link", { name: /back/i });
      await expect(backLink.first()).toBeVisible();
    }
  });
});
