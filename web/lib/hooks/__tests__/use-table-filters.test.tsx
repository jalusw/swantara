import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { TableFilterConfig } from "../use-table-filters";
import { useTableFilters } from "../use-table-filters";

const config: TableFilterConfig[] = [
  { id: "search", label: "Search", type: "text", placeholder: "Search…" },
  {
    id: "status",
    label: "Status",
    type: "select",
    columnId: "active",
    options: [
      { value: "true", label: "Active" },
      { value: "false", label: "Inactive" },
    ],
  },
];

describe("useTableFilters", () => {
  it("returns empty initial state", () => {
    const { result } = renderHook(() => useTableFilters(config));
    expect(result.current.columnFilters).toEqual([]);
    expect(result.current.globalFilter).toBe("");
    expect(result.current.activeCount).toBe(0);
  });

  it("updates globalFilter when text filter changes", () => {
    const { result } = renderHook(() => useTableFilters(config));
    act(() => {
      result.current.onGlobalFilterChange("alpha");
    });
    expect(result.current.globalFilter).toBe("alpha");
  });

  it("updates columnFilters when select filter changes", () => {
    const { result } = renderHook(() => useTableFilters(config));
    act(() => {
      result.current.onColumnFiltersChange([{ id: "active", value: "true" }]);
    });
    expect(result.current.columnFilters).toEqual([{ id: "active", value: "true" }]);
  });

  it("resolves columnId for select filter", () => {
    const { result } = renderHook(() => useTableFilters(config));
    act(() => {
      result.current.onColumnFiltersChange([{ id: "active", value: "false" }]);
    });
    expect(result.current.columnFilters).toHaveLength(1);
    expect(result.current.columnFilters[0]?.id).toBe("active");
    expect(result.current.activeCount).toBe(1);
  });

  it("resetFilters clears both globalFilter and columnFilters", () => {
    const { result } = renderHook(() => useTableFilters(config));
    act(() => {
      result.current.onGlobalFilterChange("test");
      result.current.onColumnFiltersChange([{ id: "active", value: "true" }]);
    });
    expect(result.current.activeCount).toBe(2);

    act(() => {
      result.current.resetFilters();
    });
    expect(result.current.globalFilter).toBe("");
    expect(result.current.columnFilters).toEqual([]);
    expect(result.current.activeCount).toBe(0);
  });

  it("counts only non-empty filters", () => {
    const { result } = renderHook(() => useTableFilters(config));
    expect(result.current.activeCount).toBe(0);

    act(() => {
      result.current.onGlobalFilterChange("x");
    });
    expect(result.current.activeCount).toBe(1);

    act(() => {
      result.current.onColumnFiltersChange([{ id: "active", value: "true" }]);
    });
    expect(result.current.activeCount).toBe(2);
  });

  it("accepts function updater for onGlobalFilterChange", () => {
    const { result } = renderHook(() => useTableFilters(config));
    act(() => {
      result.current.onGlobalFilterChange("a");
    });
    act(() => {
      result.current.onGlobalFilterChange((prev) => `${prev}b`);
    });
    expect(result.current.globalFilter).toBe("ab");
  });

  it("accepts function updater for onColumnFiltersChange", () => {
    const { result } = renderHook(() => useTableFilters(config));
    act(() => {
      result.current.onColumnFiltersChange([{ id: "active", value: "true" }]);
    });
    act(() => {
      result.current.onColumnFiltersChange((prev) =>
        prev.map((f) => (f.id === "active" ? { ...f, value: "false" } : f)),
      );
    });
    expect(result.current.columnFilters).toEqual([{ id: "active", value: "false" }]);
  });

  it("handles config with no text filter", () => {
    const selectOnly: TableFilterConfig[] = [
      {
        id: "status",
        label: "Status",
        type: "select",
        options: [{ value: "active", label: "Active" }],
      },
    ];
    const { result } = renderHook(() => useTableFilters(selectOnly));
    expect(result.current.globalFilter).toBe("");
    act(() => {
      result.current.onColumnFiltersChange([{ id: "status", value: "active" }]);
    });
    expect(result.current.columnFilters).toEqual([{ id: "status", value: "active" }]);
  });
});
