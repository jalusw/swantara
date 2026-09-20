import { describe, expect, it } from "vitest";
import { isJson, jsonContentType } from "@/lib/utils/http";

describe("jsonContentType", () => {
  it("equals application/json", () => {
    expect(jsonContentType).toBe("application/json");
  });
});

describe("isJson", () => {
  it("returns true for application/json", () => {
    expect(isJson("application/json")).toBe(true);
  });

  it("ignores charset parameters", () => {
    expect(isJson("application/json; charset=utf-8")).toBe(true);
  });

  it("returns false for non-json content types", () => {
    expect(isJson("text/html")).toBe(false);
  });

  it("returns false for null", () => {
    expect(isJson(null)).toBe(false);
  });
});
