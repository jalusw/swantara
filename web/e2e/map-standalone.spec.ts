import { test } from "@playwright/test";

test("standalone mapbox-gl OSM in headless", async ({ page }) => {
  await page.setContent(`<!doctype html><html><head>
    <link href="https://unpkg.com/mapbox-gl@3.28.1/dist/mapbox-gl.css" rel="stylesheet">
    <style>html,body{margin:0}#map{width:900px;height:480px}</style>
  </head><body><div id="map"></div>
  <script>
  import("https://unpkg.com/mapbox-gl@3.28.1/dist/mapbox-gl.min.js").then((mod) => {
    const mapbox = mod.default;
    mapbox.config.REQUIRE_ACCESS_TOKEN = false;
    const map = new mapbox.Map({
      container: document.getElementById("map"),
      style: {
        version: 8,
        sources: {
          osm: { type: "raster", tiles: ["https://tile.openstreetmap.org/{z}/{x}/{y}.png"], tileSize: 256, attribution: "© OpenStreetMap contributors", maxzoom: 19 },
        },
        layers: [{ id: "osm", type: "raster", source: "osm" }],
      },
      center: [106.8456, -6.2088],
      zoom: 11,
      testMode: true,
    });
    window.__map = map;
    map.on("load", () => window.__loaded = true);
  });
  </script></body></html>`);

  await page
    .waitForFunction("window.__map && window.__loaded", undefined, {
      timeout: 30000,
    })
    .catch(() => console.log("map load event did not fire"));

  await page.waitForTimeout(6000);
  const state = await page.evaluate(() => {
    const m = (window as Window & { __map?: unknown }).__map;
    if (!m) return { noMap: true };
    const map = m as {
      loaded: () => boolean;
      isStyleLoaded: () => boolean;
      isSourceLoaded: (id: string) => boolean;
      getZoom: () => number;
      getStyle: () => { layers: Array<{ id: string }> };
    };
    return {
      loaded: map.loaded(),
      isStyleLoaded: map.isStyleLoaded(),
      isSourceLoaded: map.isSourceLoaded("osm"),
      zoom: map.getZoom(),
      layers: map.getStyle().layers.map((l) => l.id),
    };
  });
  console.log("standalone state:", JSON.stringify(state));

  await page.locator("#map").screenshot({ path: "/tmp/opencode/standalone.png" });
});
