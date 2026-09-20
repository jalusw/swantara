import { describe, expect, it } from "vitest";
import { getColumnLabel, getColumnValue, resolvePath } from "../table";

describe("getColumnLabel", () => {
  it("prefers the string header", () => {
    expect(getColumnLabel({ header: "Name" })).toBe("Name");
  });

  it("falls back to accessorKey", () => {
    expect(getColumnLabel({ accessorKey: "email" })).toBe("email");
  });

  it("falls back to id and then empty string", () => {
    expect(getColumnLabel({ id: "actions" } as never)).toBe("actions");
    expect(getColumnLabel({} as never)).toBe("");
  });
});

describe("resolvePath", () => {
  it("resolves nested paths", () => {
    expect(resolvePath({ a: { b: 2 } }, "a.b")).toBe(2);
  });

  it("returns undefined for missing or null segments", () => {
    expect(resolvePath({ a: {} }, "a.b.c")).toBeUndefined();
    expect(resolvePath(null, "a")).toBeUndefined();
  });
});

describe("getColumnValue", () => {
  it("uses accessorFn when present", () => {
    expect(
      getColumnValue({ id: "n", accessorFn: (row: { n: number }) => row.n * 2 }, { n: 3 }),
    ).toBe(6);
  });

  it("resolves accessorKey paths", () => {
    expect(getColumnValue({ accessorKey: "a.b" }, { a: { b: "hit" } })).toBe("hit");
  });

  it("returns undefined without an accessor", () => {
    expect(getColumnValue({ id: "actions" } as never, { a: 1 })).toBeUndefined();
  });
});
