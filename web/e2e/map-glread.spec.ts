import { test } from "@playwright/test";

test("read webgl pixels from map canvas", async ({ page }) => {
  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(20000);

  const result = await page.evaluate(() => {
    return Array.from(document.querySelectorAll("canvas")).map((c, idx) => {
      const canvas = c as HTMLCanvasElement;
      let gl: WebGL2RenderingContext | WebGLRenderingContext | null = null;
      try {
        gl = canvas.getContext("webgl2") || canvas.getContext("webgl");
      } catch {
        gl = null;
      }
      if (!gl) return { idx, state: "no-gl" };
      const w = gl.drawingBufferWidth;
      const h = gl.drawingBufferHeight;
      if (w === 0 || h === 0) return { idx, state: "zero-size", w, h };
      const data = new Uint8Array(w * h * 4);
      try {
        gl.readPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, data);
      } catch (e) {
        return { idx, state: `readPixels-error:${(e as Error).message}`, w, h };
      }
      let lit = 0;
      let total = 0;
      const colors = new Map<string, number>();
      for (let i = 0; i < data.length; i += 4) {
        const r = data[i]!,
          g = data[i + 1]!,
          b = data[i + 2]!,
          a = data[i + 3]!;
        total++;
        const lum = 0.2126 * r + 0.7152 * g + 0.0722 * b;
        if (lum > 40 && a > 0) lit++;
        const key = `${r & 0xf8},${g & 0xf8},${b & 0xf8}`;
        colors.set(key, (colors.get(key) || 0) + 1);
      }
      const top = [...colors.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5);
      return { idx, w, h, lit, total, top };
    });
  });

  console.log(JSON.stringify(result, null, 2));
});
