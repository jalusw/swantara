import { describe, expect, it } from "vitest";
import { logger } from "../logger";

describe("server logger extra", () => {
  it("emits an info log without throwing", () => {
    expect(() => logger.info("extra info")).not.toThrow();
  });

  it("emits a warn log with context without throwing", () => {
    expect(() => logger.warn({ route: "/api/v1/me" }, "extra warn")).not.toThrow();
  });

  it("emits an error log with an Error without throwing", () => {
    expect(() => logger.error({ err: new Error("boom") }, "extra error")).not.toThrow();
  });

  it("emits a debug log without throwing", () => {
    expect(() => logger.debug("extra debug")).not.toThrow();
  });

  it("creates a child logger", () => {
    const child = logger.child({ scope: "extra" });
    expect(child).toBeDefined();
    expect(() => child.info("child info")).not.toThrow();
  });
});
