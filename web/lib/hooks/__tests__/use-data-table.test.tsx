import { renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { tableStateToListQuery, useDataTableState } from "../use-data-table";

describe("useDataTableState", () => {
  it("returns default pagination", () => {
    const { result } = renderHook(() => useDataTableState());
    expect(result.current.pagination).toEqual({ pageIndex: 0, pageSize: 10 });
  });

  it("accepts initial page size", () => {
    const { result } = renderHook(() => useDataTableState({ initialPageSize: 25 }));
    expect(result.current.pagination.pageSize).toBe(25);
  });

  it("resetFilters clears global filter and column filters", () => {
    const { result } = renderHook(() => useDataTableState());
    result.current.setGlobalFilter("test");
    result.current.resetFilters();
    expect(result.current.globalFilter).toBe("");
    expect(result.current.columnFilters).toEqual([]);
  });
});

describe("tableStateToListQuery", () => {
  it("converts pagination to 1-based page", () => {
    const q = tableStateToListQuery({ pagination: { pageIndex: 2, pageSize: 10 } });
    expect(q.page).toBe(3);
    expect(q.size).toBe(10);
  });

  it("converts sorting to query string", () => {
    const q = tableStateToListQuery({ sorting: [{ id: "name", desc: true }] });
    expect(q.sort).toBe("name:desc");
  });

  it("builds filters from column filters", () => {
    const q = tableStateToListQuery({
      columnFilters: [{ id: "status", value: "active" }],
    });
    expect(q.filter).toContain("status:eq:active");
  });
});
