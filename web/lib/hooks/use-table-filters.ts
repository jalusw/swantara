"use client";

import type { ColumnFiltersState, OnChangeFn } from "@tanstack/react-table";
import { useCallback, useMemo, useRef, useState } from "react";
import type { FilterConfig, FiltersMap, FilterValue } from "@/components/filter-bar";

export type TableFilterConfig = FilterConfig & {
  columnId?: string;
};

function isFilterActive(value: FilterValue): boolean {
  if (value == null || value === "") {
    return false;
  }
  if (typeof value === "object" && "from" in (value as Record<string, unknown>)) {
    return (value as { from?: unknown })?.from != null;
  }
  return true;
}

function toColumnFilters(
  config: readonly TableFilterConfig[],
  filters: FiltersMap,
): ColumnFiltersState {
  const state: ColumnFiltersState = [];
  for (const filter of config) {
    if (filter.type === "text" && filter.id === "search") {
      continue;
    }
    const value = filters[filter.id];
    if (isFilterActive(value as FilterValue)) {
      state.push({ id: filter.columnId ?? filter.id, value });
    }
  }
  return state;
}

function toGlobalFilter(config: readonly TableFilterConfig[], filters: FiltersMap): string {
  const global = config.find((f) => f.type === "text" && f.id === "search");
  if (global) {
    const value = filters[global.id];
    if (typeof value === "string") {
      return value;
    }
    return "";
  }
  for (const filter of config) {
    if (filter.type === "text") {
      const value = filters[filter.id];
      if (typeof value === "string") {
        return value;
      }
    }
  }
  return "";
}

export function useTableFilters(config: readonly TableFilterConfig[]): {
  columnFilters: ColumnFiltersState;
  onColumnFiltersChange: OnChangeFn<ColumnFiltersState>;
  globalFilter: string;
  onGlobalFilterChange: OnChangeFn<string>;
  resetFilters: () => void;
  activeCount: number;
} {
  const empty = useMemo(() => {
    const base: FiltersMap = {};
    for (const filter of config) {
      base[filter.id] = null;
    }
    return base;
  }, [config]);

  const [filters, setFilters] = useState<FiltersMap>({ ...empty });

  const columnFilters = useMemo(() => toColumnFilters(config, filters), [config, filters]);
  const globalFilter = useMemo(() => toGlobalFilter(config, filters), [config, filters]);

  const columnFiltersRef = useRef(columnFilters);
  columnFiltersRef.current = columnFilters;

  const globalFilterRef = useRef(globalFilter);
  globalFilterRef.current = globalFilter;

  const onColumnFiltersChange: OnChangeFn<ColumnFiltersState> = useCallback(
    (updater) => {
      const next =
        typeof updater === "function"
          ? (updater as (old: ColumnFiltersState) => ColumnFiltersState)(columnFiltersRef.current)
          : updater;
      setFilters((prev) => {
        const merged = { ...prev };
        for (const filter of config) {
          if (filter.type === "text" && filter.id === "search") {
            continue;
          }
          const columnId = filter.columnId ?? filter.id;
          const cf = next.find((f) => f.id === columnId);
          merged[filter.id] = (cf?.value as FilterValue) ?? null;
        }
        return merged;
      });
    },
    [config],
  );

  const onGlobalFilterChange: OnChangeFn<string> = useCallback(
    (updater) => {
      const next =
        typeof updater === "function"
          ? (updater as (old: string) => string)(globalFilterRef.current)
          : updater;
      setFilters((prev) => {
        const merged = { ...prev };
        const global = config.find((f) => f.type === "text" && f.id === "search");
        if (global) {
          merged[global.id] = next || null;
        } else {
          for (const filter of config) {
            if (filter.type === "text") {
              merged[filter.id] = next || null;
            }
          }
        }
        return merged;
      });
    },
    [config],
  );

  const resetFilters = useCallback(() => {
    setFilters({ ...empty });
  }, [empty]);

  const activeCount = useMemo(() => {
    return Object.values(filters).filter((value) => isFilterActive(value as FilterValue)).length;
  }, [filters]);

  return {
    columnFilters,
    onColumnFiltersChange,
    globalFilter,
    onGlobalFilterChange,
    resetFilters,
    activeCount,
  };
}
