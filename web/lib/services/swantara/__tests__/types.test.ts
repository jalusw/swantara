import { describe, expect, it } from "vitest";
import * as types from "../types";

describe("swantara types", () => {
  it("imports without error", () => {
    expect(types).toBeDefined();
  });
});
