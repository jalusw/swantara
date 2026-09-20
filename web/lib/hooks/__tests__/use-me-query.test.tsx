import { describe, expect, it } from "vitest";
import { renderHookWithProviders } from "@/lib/tests";
import { useMeOrganizationsQuery, useMeQuery } from "../use-me-query";

describe("useMeQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useMeQuery());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("returns a query result with correct shape when disabled", () => {
    const { result } = renderHookWithProviders(() => useMeQuery({ enabled: false }));
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});

describe("useMeOrganizationsQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useMeOrganizationsQuery());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("does not fetch when disabled", () => {
    const { result } = renderHookWithProviders(() => useMeOrganizationsQuery({ enabled: false }));
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});
