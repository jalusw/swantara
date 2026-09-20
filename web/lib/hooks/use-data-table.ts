"use client";

import type {
  ColumnFiltersState,
  PaginationState,
  SortingState,
  VisibilityState,
} from "@tanstack/react-table";
import { useCallback, useState } from "react";
import type { ListQuery } from "@/lib/services/swantara";

export type DataTableQueryState = {
  pagination: PaginationState;
  sorting: SortingState;
  globalFilter: string;
  columnFilters: ColumnFiltersState;
  columnVisibility: VisibilityState;
};

export type UseDataTableOptions = {
  initialPageSize?: number;
  initialSorting?: SortingState;
  initialGlobalFilter?: string;
  initialColumnFilters?: ColumnFiltersState;
  initialColumnVisibility?: VisibilityState;
};

export function useDataTableState(options: UseDataTableOptions = {}) {
  const {
    initialPageSize = 10,
    initialSorting = [],
    initialGlobalFilter = "",
    initialColumnFilters = [],
    initialColumnVisibility = {},
  } = options;

  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: initialPageSize,
  });
  const [sorting, setSorting] = useState<SortingState>(initialSorting);
  const [globalFilter, setGlobalFilter] = useState(initialGlobalFilter);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>(initialColumnFilters);
  const [columnVisibility, setColumnVisibility] =
    useState<VisibilityState>(initialColumnVisibility);

  const resetFilters = useCallback(() => {
    setGlobalFilter("");
    setColumnFilters([]);
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  }, []);

  const resetSorting = useCallback(() => {
    setSorting([]);
  }, []);

  return {
    pagination,
    setPagination,
    sorting,
    setSorting,
    globalFilter,
    setGlobalFilter,
    columnFilters,
    setColumnFilters,
    columnVisibility,
    setColumnVisibility,
    resetFilters,
    resetSorting,
  };
}

export type ListQueryOptions = {
  pagination?: PaginationState;
  sorting?: SortingState;
  globalFilter?: string;
  columnFilters?: ColumnFiltersState;
  searchKeys?: string[];
  globalFilterOperator?: "like" | "eq" | "in";
};

export function tableStateToListQuery(
  state: Partial<DataTableQueryState> & ListQueryOptions,
): ListQuery {
  const {
    pagination,
    sorting,
    globalFilter,
    columnFilters,
    searchKeys,
    globalFilterOperator = "like",
  } = state;

  const query: ListQuery = {};

  if (pagination) {
    query.page = pagination.pageIndex + 1;
    query.size = pagination.pageSize;
  }

  if (sorting && sorting.length > 0) {
    query.sort = sorting.map((s) => `${s.id}:${s.desc ? "desc" : "asc"}`).join(",");
  }

  const filters: string[] = [];

  if (columnFilters) {
    for (const f of columnFilters) {
      if (f.value == null || f.value === "") {
        continue;
      }
      const value = String(f.value);
      filters.push(`${String(f.id)}:eq:${value}`);
    }
  }

  if (globalFilter && searchKeys && searchKeys.length > 0) {
    const value = globalFilterOperator === "like" ? `%${globalFilter}%` : globalFilter;
    if (searchKeys.length === 1) {
      filters.push(`${searchKeys[0]}:${globalFilterOperator}:${value}`);
    } else {
      for (const key of searchKeys) {
        filters.push(`${key}:${globalFilterOperator}:${value}`);
      }
    }
  }

  if (filters.length > 0) {
    query.filter = filters;
  }

  return query;
}
