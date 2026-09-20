import { act } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { type FilterConfig, useFilters } from "@/components/filter-bar";
import { renderHookWithProviders } from "@/lib/tests";

const config: FilterConfig[] = [
  { id: "status", label: "Status", type: "select", options: [] },
  { id: "search", label: "Search", type: "text" },
  { id: "range", label: "Range", type: "date-range" },
];

describe("useFilters", () => {
  it("starts empty", () => {
    const { result } = renderHookWithProviders(() => useFilters(config));
    expect(result.current.activeCount).toBe(0);
  });

  it("tracks activated filters", () => {
    const { result } = renderHookWithProviders(() => useFilters(config));
    act(() => result.current.setFilter("status", "paid"));
    act(() => result.current.setFilter("search", "invoice"));
    expect(result.current.activeCount).toBe(2);
    expect(result.current.filters.search).toBe("invoice");
  });

  it("resets all filters", () => {
    const { result } = renderHookWithProviders(() => useFilters(config));
    act(() => result.current.setFilter("status", "paid"));
    act(() => result.current.resetFilters());
    expect(result.current.activeCount).toBe(0);
  });
});
