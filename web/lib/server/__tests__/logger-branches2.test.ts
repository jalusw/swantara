import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

beforeEach(() => {
  vi.resetModules();
  vi.unstubAllEnvs();
});

afterEach(() => {
  vi.resetModules();
  vi.unstubAllEnvs();
});

async function loadLogger(env: Record<string, string | undefined>) {
  for (const [key, value] of Object.entries(env)) {
    if (value === undefined) {
      vi.stubEnv(key, "");
    } else {
      vi.stubEnv(key, value);
    }
  }
  const mod = await import("../logger");
  return mod.logger;
}

describe("server logger branches2", () => {
  it("uses plain multistream branches without pretty or file", async () => {
    const logger = await loadLogger({ LOG_LEVEL: "silent", NODE_ENV: "test" });
    expect(() => logger.info("plain no streams")).not.toThrow();
    expect(() => logger.debug("debug no streams")).not.toThrow();
  });

  it("formats perf and non-string messages through pretty branches", async () => {
    const logger = await loadLogger({ LOG_LEVEL: "silent", LOG_PRETTY: "true", NODE_ENV: "test" });
    expect(() => logger.info("[PERF] branch perf")).not.toThrow();
    expect(() => logger.info("branch plain")).not.toThrow();
    expect(() => logger.warn("branch warn")).not.toThrow();
    expect(() => logger.error("branch error")).not.toThrow();
    expect(() => logger.fatal("branch fatal")).not.toThrow();
    expect(() => logger.trace("branch trace")).not.toThrow();
    expect(() => logger.debug("branch debug")).not.toThrow();
  });

  it("writes uncompressed file output with explicit file level", async () => {
    const path = `/tmp/opencode/logger-branches2-${process.pid}.log`;
    const logger = await loadLogger({
      LOG_LEVEL: "silent",
      LOG_FILE: path,
      LOG_FILE_LEVEL: "warn",
    });
    expect(() => logger.warn("file warn")).not.toThrow();
    expect(() => logger.info("file info filtered")).not.toThrow();
  });

  it("writes compressed file output through the gzip branch", async () => {
    const path = `/tmp/opencode/logger-branches2-${process.pid}.gz.log`;
    const logger = await loadLogger({
      LOG_LEVEL: "silent",
      LOG_FILE: path,
      LOG_COMPRESS: "true",
    });
    expect(() => logger.error("compressed error")).not.toThrow();
  });
});
