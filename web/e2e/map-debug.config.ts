import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: ".",
  projects: [{ name: "chromium", use: {} }],
  use: {
    browserName: "chromium",
    launchOptions: {
      executablePath: "/usr/bin/chromium",
      args: [
        "--no-sandbox",
        "--enable-unsafe-swiftshader",
        "--use-gl=angle",
        "--use-angle=swiftshader",
        "--enable-webgl",
        "--ignore-gpu-blocklist",
        "--disable-gpu-sandbox",
      ],
    },
  },
});
