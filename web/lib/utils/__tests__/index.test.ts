import { describe, expect, it } from "vitest";
import * as utils from "../index";

describe("utils barrel exports", () => {
  it("exports utils module", () => {
    expect(utils).toBeDefined();
  });

  it("exports perf utilities", () => {
    expect(typeof utils.perfLogLevel).toBe("function");
    expect(typeof utils.formatPerfLog).toBe("function");
  });

  it("exports table utilities", () => {
    expect(typeof utils.getColumnLabel).toBe("function");
  });

  it("exports url utilities", () => {
    expect(typeof utils.isProtectedPage).toBe("function");
  });
});
