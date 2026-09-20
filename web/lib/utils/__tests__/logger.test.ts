import { describe, expect, it } from "vitest";
import { logger } from "@/lib/utils/logger";

describe("logger", () => {
  it("is defined", () => {
    expect(logger).toBeDefined();
  });

  it("exposes a level property", () => {
    expect(typeof logger.level).toBe("string");
  });
});
