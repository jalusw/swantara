import { test } from "@playwright/test";

test("debug map rendering", async ({ page }) => {
  const logs: string[] = [];
  const tileReqs: string[] = [];
  page.on("console", (m) => logs.push(`[console.${m.type()}] ${m.text()}`));
  page.on("pageerror", (e) => logs.push(`[pageerror] ${e.message}`));
  page.on("request", (r) => {
    if (r.url().includes("tiles.openfreemap.org")) loggedReqsPush(tileReqs, r.url());
    if (r.url().includes("api.mapbox.com")) loggedReqsPush(tileReqs, `MAPBOX:${r.url()}`);
    if (r.url().includes("map-sessions")) loggedReqsPush(tileReqs, `SESSION:${r.url()}`);
  });
  page.on("requestfailed", (r) =>
    loggedReqsPush(tileReqs, `FAILED:${r.url()} ${r.failure()?.errorText ?? ""}`),
  );

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(18000);

  const canvasCount = await page.locator('[data-slot="map-view"] canvas').count();
  const canvases = page.locator('[data-slot="map-view"] canvas');
  for (let i = 0; i < canvasCount; i++) {
    const c = canvases.nth(i);
    const box = await c.boundingBox();
    const hasPixels = await c.evaluate((el: HTMLCanvasElement) => {
      try {
        const ctx = el.getContext("2d");
        if (!ctx) return "no-2d";
        const d = ctx.getImageData(0, 0, el.width, el.height).data;
        let nonTransparent = 0;
        for (let j = 3; j < d.length; j += 4) if (d[j]! > 0) nonTransparent++;
        return nonTransparent;
      } catch (e) {
        return `err:${(e as Error).message}`;
      }
    });
    console.log(`CANVAS[${i}] box=${JSON.stringify(box)} paintedPixels=${hasPixels}`);
  }

  console.log("===== REQUESTS =====");
  for (const r of [...new Set(tileReqs)].slice(0, 40)) console.log(r);
  console.log("===== ERRORS =====");
  for (const l of logs) if (l.includes("error") || l.includes("Error")) console.log(l);
  console.log("canvas count", canvasCount);
});

function loggedReqsPush(arr: string[], v: string) {
  if (arr.length < 200) arr.push(v);
}
