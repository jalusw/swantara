import { describe, expect, it } from "vitest";
import { renderHookWithProviders } from "@/lib/tests";
import { useItemVariants } from "../use-item-variants";

describe("useItemVariants", () => {
  it("returns a query result with the expected shape", () => {
    const { result } = renderHookWithProviders(() => useItemVariants("1", "10"));
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("stays idle without a item id", () => {
    const { result } = renderHookWithProviders(() => useItemVariants("1", undefined));
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});
