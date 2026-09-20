import { test } from "@playwright/test";

test("force preserveDrawingBuffer + verify pixels", async ({ page }) => {
  await page.addInitScript(() => {
    const orig = HTMLCanvasElement.prototype.getContext;
    HTMLCanvasElement.prototype.getContext = function (
      this: HTMLCanvasElement,
      type: string,
      ...args: unknown[]
    ) {
      if (type === "webgl" || type === "webgl2") {
        const attrs = Object.assign({ preserveDrawingBuffer: true }, args[0] || {});
        return orig.call(this, type, attrs);
      }
      return orig.call(this, type, ...args);
    } as typeof HTMLCanvasElement.prototype.getContext;
  });

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(25000);

  const res = await page.evaluate(() => {
    const dbg = (globalThis as { __mapDebug?: { map?: unknown } }).__mapDebug;
    const map = dbg?.map as
      | { getCanvas: () => HTMLCanvasElement; loaded: () => boolean }
      | undefined;
    if (!map) return { hasMap: false };
    const canvas = map.getCanvas() as HTMLCanvasElement;
    const gl =
      canvas.getContext("webgl2") || (canvas.getContext("webgl") as WebGLRenderingContext | null);
    if (!gl) return { noGl: true };
    const w = canvas.width,
      h = canvas.height;
    const px = new Uint8Array(w * h * 4);
    gl.readPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, px);
    let lit = 0;
    const colors = new Map<string, number>();
    for (let i = 0; i < px.length; i += 4) {
      const r = px[i]!,
        g = px[i + 1]!,
        b = px[i + 2]!,
        a = px[i + 3]!;
      const lum = 0.2126 * r + 0.7152 * g + 0.0722 * b;
      if (lum > 40 && a > 0) lit++;
      const key = `${r},${g},${b},${a}`;
      colors.set(key, (colors.get(key) || 0) + 1);
    }
    return {
      w,
      h,
      lit,
      total: w * h,
      top: [...colors.entries()].sort((a, b) => b[1] - a[1]).slice(0, 8),
      loaded: map.loaded(),
    };
  });
  console.log("RESULT:", JSON.stringify(res, null, 2));
});
