import { test } from "@playwright/test";

test("capture mapbox internal events", async ({ page }) => {
  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(25000);

  const debug = await page.evaluate(() => (globalThis as { __mapDebug?: unknown }).__mapDebug);
  console.log("DEBUG:", JSON.stringify(debug, null, 2));
});
