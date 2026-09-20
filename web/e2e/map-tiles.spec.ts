import { test } from "@playwright/test";

test("map tile responses + errors", async ({ page }) => {
  const responses: Record<string, number> = {};
  const failed: string[] = [];
  const mapErrors: string[] = [];
  page.on("response", (r) => {
    if (r.url().includes("tiles.openfreemap.org")) {
      responses[r.status()] = (responses[r.status()] || 0) + 1;
    }
  });
  page.on("requestfailed", (r) => {
    if (r.url().includes("tiles.openfreemap.org"))
      failed.push(`${r.url()} :: ${r.failure()?.errorText ?? "?"}`);
  });
  page.on("console", (m) => {
    const t = m.text();
    if (/mapbox|error|fail|webgl|cors/i.test(t)) mapErrors.push(`[${m.type()}] ${t}`);
  });

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(22000);

  console.log("tile statuses:", JSON.stringify(responses));
  console.log("failed tiles:", JSON.stringify(failed.slice(0, 10)));
  console.log("map-related console:", JSON.stringify(mapErrors.slice(0, 30), null, 1));

  // Also listen for mapbox error events via a page-level hook is not easy;
  // instead report whether the OpenFreeMap style/tiles loaded w/ CORS
  const cors = await page.evaluate(async () => {
    try {
      await fetch("https://tiles.openfreemap.org/styles/positron", {
        mode: "no-cors",
      });
      return "no-cors-fetched";
    } catch (e) {
      return `fetch-failed:${(e as Error).message}`;
    }
  });
  console.log("tile fetch probe:", cors);
});
