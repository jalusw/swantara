import { describe, expect, it } from "vitest";
import { logger } from "../logger";

describe("server logger", () => {
  it("is defined", () => {
    expect(logger).toBeDefined();
  });

  it("has a configurable level", () => {
    expect(typeof logger.level).toBe("string");
  });
});
