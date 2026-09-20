import { expect, test } from "@playwright/test";

import { navigateToOrg, seedOrg } from "./fixtures/helpers";

test.describe("Plan-to-Produce flow", () => {
  test.beforeEach(async ({ page }) => {
    await seedOrg(page);
  });

  test("navigates through manufacturing screens", async ({ page }) => {
    await navigateToOrg(page, "/products/recipes");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/production-orders");
    await expect(page.getByRole("main").first()).toBeVisible();

    await navigateToOrg(page, "/products");
    await expect(page.getByRole("main").first()).toBeVisible();
  });

  test("BOMs page has expected heading", async ({ page }) => {
    await navigateToOrg(page, "/products/recipes");
    await expect(page.getByRole("main").first()).toBeVisible();
    await expect(page.getByRole("heading")).toBeVisible();
  });
});
