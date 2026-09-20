// @vitest-environment node
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

async function loadConfig() {
  vi.resetModules();
  const { serverConfig: config } = await import("../config");
  return config;
}

describe("server config", () => {
  beforeEach(() => {
    vi.stubEnv("SWANTARA_API_URL", "http://localhost:8080");
    vi.stubEnv("LOG_LEVEL", "silent");
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("throws when the API URL is not set", async () => {
    delete process.env.SWANTARA_API_URL;
    await expect(loadConfig()).rejects.toThrow("API_URL IS NOT SET");
  });

  it("uses sensible defaults for the log config", async () => {
    const config = await loadConfig();

    expect(config.serviceApiUrl).toBe("http://localhost:8080");
    expect(config.log).toEqual({
      level: "silent",
      pretty: false,
      file: undefined,
      fileLevel: undefined,
      compress: false,
    });
  });

  it("enables pretty printing in development or when LOG_PRETTY is true", async () => {
    vi.stubEnv("NODE_ENV", "development");
    let config = await loadConfig();
    expect(config.log.pretty).toBe(true);

    vi.resetModules();
    vi.stubEnv("NODE_ENV", "test");
    vi.stubEnv("LOG_PRETTY", "true");
    config = await loadConfig();
    expect(config.log.pretty).toBe(true);
  });

  it("reads the file logging options", async () => {
    vi.stubEnv("LOG_FILE", "/tmp/swantara.log");
    vi.stubEnv("LOG_FILE_LEVEL", "debug");
    vi.stubEnv("LOG_COMPRESS", "true");
    vi.stubEnv("APP_NAME", "swantara-web");

    const config = await loadConfig();
    expect(config.appName).toBe("swantara-web");
    expect(config.log.file).toBe("/tmp/swantara.log");
    expect(config.log.fileLevel).toBe("debug");
    expect(config.log.compress).toBe(true);
  });
});
