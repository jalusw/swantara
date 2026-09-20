import { test } from "@playwright/test";

test("screenshot map section", async ({ page }) => {
  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(15000);

  const map = page.locator('[data-slot="map-view"]').first();
  await map.scrollIntoViewIfNeeded();
  await page.waitForTimeout(3000);
  await page.screenshot({ path: "/tmp/opencode/map-1.png" });

  const box = await map.boundingBox();
  console.log("first map box:", JSON.stringify(box));

  const second = page.locator('[data-slot="map-view"]').nth(1);
  await second.scrollIntoViewIfNeeded();
  await page.waitForTimeout(3000);
  await page.screenshot({ path: "/tmp/opencode/map-2.png" });
  console.log("second map box:", JSON.stringify(await second.boundingBox()));

  const sectionTitles = await page.locator("main h2, main h3").allTextContents();
  console.log("headings:", JSON.stringify(sectionTitles));
});
