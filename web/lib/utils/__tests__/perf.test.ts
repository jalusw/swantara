import { describe, expect, it } from "vitest";
import { formatPerfLog, perfLogLevel } from "../perf";

describe("perfLogLevel", () => {
  it("returns error when hasError is true", () => {
    expect(perfLogLevel(100, 200, true)).toBe("error");
  });

  it("returns warn when duration exceeds threshold", () => {
    expect(perfLogLevel(1500, 200)).toBe("warn");
  });

  it("returns error for 500+ status", () => {
    expect(perfLogLevel(100, 500)).toBe("error");
  });

  it("returns info for fast successful requests", () => {
    expect(perfLogLevel(100, 200)).toBe("info");
  });
});

describe("formatPerfLog", () => {
  it("formats a perf log line", () => {
    const result = formatPerfLog("GET", "/api/users", 200, "123ms");
    expect(result).toBe("[PERF] GET /api/users → 200 (123ms)");
  });

  it("formats error outcome", () => {
    const result = formatPerfLog("POST", "/api/data", "ERROR", "50ms");
    expect(result).toBe("[PERF] POST /api/data → ERROR (50ms)");
  });
});
