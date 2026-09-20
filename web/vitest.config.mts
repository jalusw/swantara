import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const rootDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  test: {
    environment: "jsdom",
    setupFiles: ["./test-setup.ts"],
    testTimeout: 15_000,
    env: {
      SWANTARA_API_URL: "http://localhost:8080",
      LOG_LEVEL: "silent",
      NODE_ENV: "test",
    },
    exclude: ["node_modules", ".next", "coverage", "playwright-report", "test-results", "e2e"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov"],
      include: [
        "app/**/*.{ts,tsx}",
        "components/**/*.{ts,tsx}",
        "lib/**/*.{ts,tsx}",
        "providers/**/*.{ts,tsx}",
        "stores/**/*.{ts,tsx}",
        "types/**/*.ts",
      ],
      exclude: ["**/*.test.{ts,tsx}", "**/__tests__/**"],
      thresholds: {
        statements: 80,
        branches: 80,
        functions: 80,
        lines: 80,
      },
    },
  },
  resolve: {
    alias: [
      {
        find: "@lib",
        replacement: path.resolve(rootDir, "./lib"),
      },
      {
        find: "@swantara/ui",
        replacement: path.resolve(rootDir, "./components"),
      },
      { find: "@", replacement: rootDir },
    ],
  },
});
