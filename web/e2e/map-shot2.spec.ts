import { test } from "@playwright/test";

test("capture map with swiftshader", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(`pageerror: ${e.message}`));
  page.on("console", (m) => {
    if (m.type() === "error" || m.type() === "warning")
      errors.push(`console.${m.type()}: ${m.text()}`);
  });

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(20000);

  const map = page.locator('[data-slot="map-view"]').first();
  await map.scrollIntoViewIfNeeded();
  await page.waitForTimeout(4000);
  await map.screenshot({ path: "/tmp/opencode/map-swiftshader.png" });

  console.log("errors:", JSON.stringify(errors.slice(0, 20), null, 1));
});
