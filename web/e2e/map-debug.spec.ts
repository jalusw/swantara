import { test } from "@playwright/test";

test("debug map console", async ({ page }) => {
  const logs: string[] = [];
  page.on("console", (m) => logs.push(`[console.${m.type()}] ${m.text()}`));
  page.on("pageerror", (e) => logs.push(`[pageerror] ${e.message}`));

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(15000);

  console.log("===== LOGS =====");
  for (const l of logs) console.log(l);

  const maps = await page.locator('[data-slot="map-view"] canvas').count();
  console.log("===== CANVAS COUNT =====", maps);
});
