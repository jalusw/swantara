import { describe, expect, it } from "vitest";
import { toCamelCase, toSnakeCase } from "@/lib/utils/case";

describe("toSnakeCase", () => {
  it("converts camelCase keys to snake_case", () => {
    expect(toSnakeCase({ firstName: "a", lastName: "b" })).toEqual({
      first_name: "a",
      last_name: "b",
    });
  });

  it("returns primitives unchanged", () => {
    expect(toSnakeCase("hello")).toBe("hello");
    expect(toSnakeCase(42)).toBe(42);
  });

  it("returns null and undefined unchanged", () => {
    expect(toSnakeCase(null)).toBeNull();
    expect(toSnakeCase(undefined)).toBeUndefined();
  });
});

describe("toCamelCase", () => {
  it("converts snake_case keys to camelCase", () => {
    expect(toCamelCase({ first_name: "a", last_name: "b" })).toEqual({
      firstName: "a",
      lastName: "b",
    });
  });

  it("returns primitives unchanged", () => {
    expect(toCamelCase("hello")).toBe("hello");
    expect(toCamelCase(42)).toBe(42);
  });

  it("converts nested objects deeply", () => {
    expect(toCamelCase({ user_data: { created_at: 1 } })).toEqual({
      userData: { createdAt: 1 },
    });
  });
});
