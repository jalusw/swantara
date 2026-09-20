import { test } from "@playwright/test";

test("query live map state", async ({ page }) => {
  const consoleLines: string[] = [];
  page.on("console", (m) => {
    if (m.text().includes("map-debug")) consoleLines.push(m.text());
  });

  await page.goto("http://localhost:3000/preview/design-system", {
    waitUntil: "domcontentloaded",
    timeout: 90000,
  });
  await page.waitForTimeout(25000);

  const state = await page.evaluate(() => {
    const dbg = (globalThis as { __mapDebug?: { map?: unknown } }).__mapDebug;
    const map = dbg?.map as
      | {
          loaded: () => boolean;
          isStyleLoaded: () => boolean;
          isSourceLoaded: (id: string) => boolean;
          getCanvas: () => HTMLCanvasElement;
          getContainer: () => HTMLElement;
          getZoom: () => number;
          getStyle: () => { layers: Array<{ id: string }> };
          accessToken: string;
        }
      | undefined;
    if (!map) return { hasMap: false };
    let res: Record<string, unknown>;
    try {
      res = {
        loaded: map.loaded(),
        isStyleLoaded: map.isStyleLoaded(),
        isSourceLoaded: map.isSourceLoaded("openmaptiles"),
        hasCanvas: !!map.getCanvas(),
        canvasSize:
          map.getCanvas().getBoundingClientRect().width +
          "x" +
          map.getCanvas().getBoundingClientRect().height,
        getContainerClientWidth: map.getContainer().clientWidth,
        getContainerClientHeight: map.getContainer().clientHeight,
        zoom: map.getZoom(),
        listLayers: map.getStyle().layers.map((l) => l.id),
        hasAccessToken: typeof map.accessToken === "string" && map.accessToken.length > 0,
      };
    } catch (e) {
      res = { queryError: (e as Error).message };
    }
    return { hasMap: true, res };
  });

  console.log("MAP STATE:", JSON.stringify(state, null, 2));
  console.log("MAP DEBUG LOGS:", JSON.stringify(consoleLines, null, 1));
});
