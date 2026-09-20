"use client";

import type { ColumnFiltersState, OnChangeFn } from "@tanstack/react-table";
import { useMemo } from "react";
import type { TableFilterConfig } from "@/lib/hooks/use-table-filters";
import { cn } from "@/lib/utils";
import { FilterBar, type FiltersMap, type FilterValue } from "./filter-bar";

export type TableFilterBarProps = {
  config: readonly TableFilterConfig[];
  columnFilters: ColumnFiltersState;
  onColumnFiltersChange: OnChangeFn<ColumnFiltersState>;
  globalFilter?: string;
  onGlobalFilterChange?: OnChangeFn<string>;
  onReset?: () => void;
  className?: string;
  actions?: React.ReactNode;
};

function isFilterActive(value: FilterValue): boolean {
  if (value == null || value === "") return false;
  if (typeof value === "object" && "from" in (value as Record<string, unknown>)) {
    return (value as { from?: unknown })?.from != null;
  }
  return true;
}

function toFiltersMap(
  config: readonly TableFilterConfig[],
  columnFilters: ColumnFiltersState,
  globalFilter: string,
): FiltersMap {
  const map: FiltersMap = {};
  for (const filter of config) {
    if (filter.type === "text" && filter.id === "search") {
      map[filter.id] = globalFilter || null;
    } else if (filter.type === "text") {
      const columnId = filter.columnId ?? filter.id;
      const cf = columnFilters.find((f) => f.id === columnId);
      map[filter.id] = (cf?.value as FilterValue) ?? null;
    } else {
      const columnId = filter.columnId ?? filter.id;
      const cf = columnFilters.find((f) => f.id === columnId);
      map[filter.id] = (cf?.value as FilterValue) ?? null;
    }
  }
  return map;
}

export function TableFilterBar({
  config,
  columnFilters,
  onColumnFiltersChange,
  globalFilter = "",
  onGlobalFilterChange,
  onReset,
  className,
  actions,
}: TableFilterBarProps) {
  const filtersMap = useMemo(
    () => toFiltersMap(config, columnFilters, globalFilter),
    [config, columnFilters, globalFilter],
  );

  function handleFilterChange(id: string, value: FilterValue) {
    const filter = config.find((c) => c.id === id);
    if (!filter) {
      return;
    }

    if (filter.type === "text" && filter.id === "search") {
      onGlobalFilterChange?.(typeof value === "string" ? value : "");
      return;
    }

    if (filter.type === "text") {
      const columnId = filter.columnId ?? id;
      onColumnFiltersChange((prev) => {
        const next = prev.filter((f) => f.id !== columnId);
        if (isFilterActive(value)) {
          next.push({ id: columnId, value });
        }
        return next;
      });
      return;
    }

    const columnId = filter.columnId ?? id;
    onColumnFiltersChange((prev) => {
      const next = prev.filter((f) => f.id !== columnId);
      if (isFilterActive(value)) {
        next.push({ id: columnId, value });
      }
      return next;
    });
  }

  return (
    <div data-slot="table-filter-bar" className={cn("w-full min-w-0 flex-1", className)}>
      <FilterBar
        config={config}
        filters={filtersMap}
        onFilterChange={handleFilterChange}
        onReset={onReset}
        className="w-full"
        actions={actions}
      />
    </div>
  );
}
