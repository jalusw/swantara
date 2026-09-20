import { describe, expect, it } from "vitest";
import { renderHookWithProviders } from "@/lib/tests";
import { useItemVariantsQuery, useProductQuery } from "../use-item-query";

describe("useProductQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useProductQuery("1", "10"));
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("is disabled when orgId is empty", () => {
    const { result } = renderHookWithProviders(() => useProductQuery("", "10"));
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});

describe("useItemVariantsQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useItemVariantsQuery("1", "10"));
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("is disabled when itemId is empty", () => {
    const { result } = renderHookWithProviders(() => useItemVariantsQuery("1", ""));
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});
