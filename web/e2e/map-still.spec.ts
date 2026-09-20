import { test } from "@playwright/test";

test("inspect map canvas deep", async ({ page }) => {
  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(20000);

  const mapContainers = page.locator('[data-slot="map-view"]');
  const n = await mapContainers.count();
  console.log("map containers:", n);

  for (let i = 0; i < n; i++) {
    const mc = mapContainers.nth(i);
    await mc.scrollIntoViewIfNeeded();
    await page.waitForTimeout(2500);
    await mc.screenshot({ path: `/tmp/opencode/map-el-${i}.png` });

    const info = await page.evaluate(() => {
      const all = Array.from(document.querySelectorAll("canvas"));
      return all.map((c) => {
        const canvas = c as HTMLCanvasElement;
        let ctxState = "none";
        for (const t of ["webgl2", "webgl", "experimental-webgl"] as const) {
          const gl = canvas.getContext(t);
          if (gl) {
            ctxState = `${t}:ok`;
            break;
          }
        }
        const rect = canvas.getBoundingClientRect();
        return {
          ctx: ctxState,
          width: canvas.width,
          height: canvas.height,
          css: `${rect.width}x${rect.height}`,
          visible: rect.width > 0 && rect.height > 0,
        };
      });
    });
    console.log(`map[${i}] canvases:`, JSON.stringify(info));
  }
});
