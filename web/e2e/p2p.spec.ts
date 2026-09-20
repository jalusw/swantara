import { expect, test } from "@playwright/test";

import { navigateToOrg, seedOrg } from "./fixtures/helpers";

test.describe("Procure-to-Pay (P2P) flow", () => {
  test.beforeEach(async ({ page }) => {
    await seedOrg(page);
  });

  test("navigates through P2P screens", async ({ page }) => {
    await navigateToOrg(page, "/purchases/requisitions");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/purchases");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/stock/shipments");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/invoices");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/payments");
    await expect(page.getByRole("main").first()).toBeVisible();
  });

  test("purchase orders page has expected structure", async ({ page }) => {
    await navigateToOrg(page, "/purchases");
    await expect(page.getByRole("main").first()).toBeVisible();
    await expect(page.getByRole("heading")).toBeVisible();
  });
});
