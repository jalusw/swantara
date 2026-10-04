"use client";

import type { ColumnDef, ColumnFiltersState, OnChangeFn } from "@tanstack/react-table";
import { DownloadIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import type { ReactNode } from "react";
import { useMemo } from "react";
import { Button } from "@/components/button";
import { TableFilterBar } from "@/components/table-filter-bar";
import { DataTable, type DataTableStatus } from "@/components/tanstack-table";
import type { TableFilterConfig } from "@/lib/hooks/use-table-filters";
import { useTableFilters } from "@/lib/hooks/use-table-filters";
import { exportCsv } from "@/lib/utils";
import { getColumnLabel, getColumnValue } from "@/lib/utils/table";

export type EntityTableStatusOption = {
  value: string;
  label: string;
};

export type InteractiveEntityTableProps<T, V> = {
  columns: ColumnDef<T, V>[];
  data: T[];
  getRowId: (row: T, index: number) => string;
  searchKeys: string[];
  statusKey?: keyof T;
  statusOptions?: EntityTableStatusOption[];
  searchPlaceholder: string;
  filterLabel?: string;
  allLabel?: string;
  ariaLabel: string;
  pageSize?: number;
  actions?: ReactNode;
  emptyTitle?: string;
  emptyDescription?: string;
  exportFileName?: string;
  status?: DataTableStatus;
};

function buildFilterConfig(
  columns: ColumnDef<unknown, unknown>[],
  searchKeys: string[],
  statusKey?: string,
  statusOptions?: EntityTableStatusOption[],
  searchPlaceholder?: string,
  filterLabel?: string,
  searchLabel?: string,
  filterBy?: (label: string) => string,
): TableFilterConfig[] {
  const config: TableFilterConfig[] = [];

  if (searchKeys.length > 0) {
    config.push({
      id: "search",
      label: searchLabel ?? "Search",
      type: "text",
      placeholder: searchPlaceholder,
    });
  }

  if (statusKey && statusOptions && statusOptions.length > 0) {
    config.push({
      id: statusKey,
      label: filterLabel ?? statusKey,
      type: "select",
      columnId: statusKey,
      options: statusOptions.map((opt) => ({ value: opt.value, label: opt.label })),
    });
  }

  for (const col of columns) {
    const colId = String(
      (col as { accessorKey?: string; id?: string }).accessorKey ??
        (col as { id?: string }).id ??
        "",
    );
    if (!colId || colId === "actions" || colId === "search" || colId === String(statusKey)) {
      continue;
    }
    if (searchKeys.includes(colId)) {
      continue;
    }
    const label = getColumnLabel(col as ColumnDef<unknown, unknown>) || colId;
    config.push({
      id: colId,
      label,
      type: "text",
      columnId: colId,
      placeholder: filterBy ? filterBy(label.toLowerCase()) : `Filter ${label.toLowerCase()}…`,
    });
  }

  return config;
}

export function InteractiveEntityTable<T, V>({
  columns,
  data,
  getRowId,
  searchKeys = [],
  statusKey,
  statusOptions = [],
  searchPlaceholder,
  filterLabel = "",
  ariaLabel,
  pageSize = 8,
  actions,
  emptyTitle,
  emptyDescription,
  exportFileName,
  status: statusOverride,
}: InteractiveEntityTableProps<T, V>) {
  const tTables = useTranslations("Tables" as unknown as "Common");
  const tx = tTables as unknown as (
    key: string,
    values?: Record<string, string | number>,
  ) => string;
  const filterConfig = useMemo(
    () =>
      buildFilterConfig(
        columns as ColumnDef<unknown, unknown>[],
        searchKeys,
        statusKey as string | undefined,
        statusOptions.length > 0 ? statusOptions : undefined,
        searchPlaceholder,
        filterLabel,
        tx("search"),
        (label: string) => tx("filterBy", { label }),
      ),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [columns, searchKeys, statusKey, statusOptions, searchPlaceholder, filterLabel, tTables],
  );

  const filters = useTableFilters(filterConfig);

  const normalized = (filters.globalFilter ?? "").trim().toLowerCase();
  const filtered = data.filter((row) => {
    const matchesQuery =
      normalized.length === 0 ||
      searchKeys.some((key) =>
        String((row as Record<string, unknown>)[key])
          .toLowerCase()
          .includes(normalized),
      );
    const matchesColumnFilters = filters.columnFilters.every((f) => {
      if (String(f.id) === String(statusKey)) {
        return (
          String((row as Record<string, unknown>)[String(statusKey)] ?? "") === String(f.value)
        );
      }
      const cell = (row as Record<string, unknown>)[String(f.id)];
      return String(cell ?? "")
        .toLowerCase()
        .includes(String(f.value).toLowerCase());
    });
    return matchesQuery && matchesColumnFilters;
  });

  const handleColumnFiltersChange: OnChangeFn<ColumnFiltersState> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: ColumnFiltersState) => ColumnFiltersState)(filters.columnFilters)
        : updater;
    filters.onColumnFiltersChange(next);
  };

  function handleResetFilters() {
    filters.resetFilters();
  }

  return (
    <DataTable
      data={filtered}
      columns={columns}
      getRowId={getRowId}
      enableSorting
      enablePagination
      enableRowSelection
      enableColumnFilters={filterConfig.some((c) => c.id !== "search")}
      pageSizeOptions={[8, 16, 32]}
      initialPageSize={pageSize}
      ariaLabel={ariaLabel}
      status={
        statusOverride ??
        (filtered.length === 0
          ? {
              type: "empty",
              title: emptyTitle ?? tx("emptyTitle"),
              description: emptyDescription ?? tx("emptyDescription"),
            }
          : { type: "none" })
      }
      toolbar={
        <TableFilterBar
          config={filterConfig}
          columnFilters={filters.columnFilters}
          onColumnFiltersChange={handleColumnFiltersChange}
          globalFilter={filters.globalFilter}
          onGlobalFilterChange={filters.onGlobalFilterChange}
          onReset={handleResetFilters}
          className="w-full"
          actions={
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const headers = columns.map((column) => getColumnLabel(column));
                  const rows = filtered.map((row) =>
                    columns.map((column) => getColumnValue(column, row)),
                  );
                  exportCsv(exportFileName ?? ariaLabel, headers, rows);
                }}
              >
                <DownloadIcon aria-hidden />
                <span>{tx("export")}</span>
              </Button>
              {actions}
            </>
          }
        />
      }
    />
  );
}
