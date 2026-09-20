import { test } from "@playwright/test";

test("inspect webgl canvas pixels", async ({ page }) => {
  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(18000);

  const result = await page.evaluate(() => {
    const canvases = Array.from(document.querySelectorAll('[data-slot="map-view"] canvas'));
    return canvases.map((c) => {
      const canvas = c as HTMLCanvasElement;
      const gl =
        canvas.getContext("webgl") ||
        (canvas.getContext("experimental-webgl") as WebGLRenderingContext | null);
      if (!gl) return { id: canvas.id, state: "no-webgl" };
      const w = gl.drawingBufferWidth;
      const h = gl.drawingBufferHeight;
      const px = new Uint8Array(w * h * 4);
      gl.readPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, px);
      let nonZero = 0;
      const counts: Record<string, number> = {};
      for (let i = 0; i < px.length; i += 4) {
        const r = px[i]!,
          g = px[i + 1]!,
          b = px[i + 2]!,
          a = px[i + 3]!;
        const key = `${r},${g},${b},${a}`;
        counts[key] = (counts[key] || 0) + 1;
        if (r + g + b + a > 0) nonZero++;
      }
      const top = Object.entries(counts)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5);
      const sw = canvas.style.width;
      const sh = canvas.style.height;
      const rect = c.getBoundingClientRect();
      return {
        id: c.id,
        w,
        h,
        style: `${sw}x${sh}`,
        rect: `${rect.width}x${rect.height}`,
        nonZero,
        total: w * h,
        topColors: top,
      };
    });
  });

  console.log(JSON.stringify(result, null, 2));
});
