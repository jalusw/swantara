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

describe("server logger branches", () => {
  it("emits every level without throwing", async () => {
    const logger = await loadLogger({ LOG_LEVEL: "silent" });
    expect(() => logger.info("branch info")).not.toThrow();
    expect(() => logger.warn("branch warn")).not.toThrow();
    expect(() => logger.error("branch error")).not.toThrow();
    expect(() => logger.debug("branch debug")).not.toThrow();
    expect(() => logger.fatal("branch fatal")).not.toThrow();
    expect(() => logger.trace("branch trace")).not.toThrow();
  });

  it("logs non-serializable and circular arguments", async () => {
    const logger = await loadLogger({ LOG_LEVEL: "silent" });
    const circular: Record<string, unknown> = { name: "loop" };
    circular.self = circular;
    expect(() => logger.info({ payload: circular }, "circular")).not.toThrow();
    expect(() => logger.info({ big: BigInt(10) }, "bigint")).not.toThrow();
    expect(() => logger.info(undefined, "undefined arg")).not.toThrow();
    expect(() => logger.error(new Error("direct error"))).not.toThrow();
  });

  it("creates child loggers with bindings", async () => {
    const logger = await loadLogger({ LOG_LEVEL: "silent" });
    const child = logger.child({ scope: "branches", requestId: "r-1" });
    expect(() => child.info("child message")).not.toThrow();
    expect(() => child.child({ nested: true }).warn("nested child")).not.toThrow();
  });

  it("respects development pretty printing with perf markers", async () => {
    const logger = await loadLogger({ LOG_LEVEL: "silent", NODE_ENV: "development" });
    expect(() => logger.info("[PERF] request took 12ms")).not.toThrow();
    expect(() => logger.info("plain message")).not.toThrow();
    expect(() => logger.warn("pretty warn")).not.toThrow();
    expect(() => logger.error("pretty error")).not.toThrow();
  });

  it("enables pretty printing through LOG_PRETTY", async () => {
    const logger = await loadLogger({
      LOG_LEVEL: "silent",
      LOG_PRETTY: "true",
      NODE_ENV: "test",
    });
    expect(() => logger.info("[PERF] op finished")).not.toThrow();
    expect(() => logger.debug("pretty debug")).not.toThrow();
  });

  it("writes to a file stream without compression", async () => {
    const path = `/tmp/opencode/logger-branches-${process.pid}.log`;
    const logger = await loadLogger({ LOG_LEVEL: "silent", LOG_FILE: path });
    expect(() => logger.info("file info")).not.toThrow();
    expect(() => logger.error("file error")).not.toThrow();
  });

  it("writes to a compressed file stream with file level", async () => {
    const path = `/tmp/opencode/logger-branches-${process.pid}.gz.log`;
    const logger = await loadLogger({
      LOG_LEVEL: "silent",
      LOG_FILE: path,
      LOG_COMPRESS: "true",
      LOG_FILE_LEVEL: "error",
    });
    expect(() => logger.info("compressed info")).not.toThrow();
    expect(() => logger.error("compressed error")).not.toThrow();
  });

  it("falls back to configured level for the file stream", async () => {
    const path = `/tmp/opencode/logger-branches-${process.pid}-fallback.log`;
    const logger = await loadLogger({
      LOG_LEVEL: "silent",
      LOG_FILE: path,
      NODE_ENV: "development",
    });
    expect(() => logger.warn("fallback warn")).not.toThrow();
  });
});
