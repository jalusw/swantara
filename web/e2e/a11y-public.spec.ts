import { expect, test } from "@playwright/test";

import { expectAccessible } from "./a11y";

const publicRoutes = [{ path: "", label: "landing" }] as const;

test.describe("public", () => {
  for (const { path, label } of publicRoutes) {
    test(`${label} has no a11y violations`, async ({ page }) => {
      await page.goto(`${path}`);

      await expect(page.getByRole("main")).toBeVisible();

      await expectAccessible(page, { ruleTags: ["wcag2a", "wcag2aa"] });
    });
  }
});
