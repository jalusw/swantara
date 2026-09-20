import { test } from "@playwright/test";

test("webgl renderer + mapbox state", async ({ page }) => {
  const logs: string[] = [];
  page.on("console", (m) => {
    if (/error|fail|webgl|cors|canvas/i.test(m.text())) logs.push(`[${m.type()}] ${m.text()}`);
  });
  page.on("pageerror", (e) => logs.push(`[pageerror] ${e.message}`));

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(25000);

  const info = await page.evaluate(() => {
    const canvas = Array.from(document.querySelectorAll("canvas")).find((c) => {
      const gl = (c as HTMLCanvasElement).getContext("webgl2");
      return !!gl;
    }) as HTMLCanvasElement | undefined;
    if (!canvas) return { ok: false };

    const gl = canvas.getContext("webgl2", {
      failIfMajorPerformanceCaveat: false,
    });
    if (!gl) return { ok: false };

    const dbg = gl.getExtension("WEBGL_debug_renderer_info");
    return {
      ok: true,
      renderer: dbg ? gl.getParameter(dbg.UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER),
      supplier: dbg ? gl.getParameter(dbg.UNMASKED_VENDOR_WEBGL) : gl.getParameter(gl.VENDOR),
      glVersion: gl.getParameter(gl.VERSION),
      shaders: gl.getParameter(gl.MAX_VERTEX_ATTRIBS),
      contextLost: gl.isContextLost(),
    };
  });

  console.log("GL INFO:", JSON.stringify(info));
  console.log("LOGS:", JSON.stringify(logs.slice(0, 20), null, 1));
});
