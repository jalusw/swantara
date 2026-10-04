"use client";

import type {
  ColumnDef,
  ColumnFiltersState,
  OnChangeFn,
  SortingState,
} from "@tanstack/react-table";
import { DownloadIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import { type ReactNode, useEffect, useMemo } from "react";
import { Button } from "@/components/button";
import { TableFilterBar } from "@/components/table-filter-bar";
import { DataTable, type DataTableStatus } from "@/components/tanstack-table";
import { tableStateToListQuery, useDataTableState } from "@/lib/hooks/use-data-table";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { TableFilterConfig } from "@/lib/hooks/use-table-filters";
import { useTableFilters } from "@/lib/hooks/use-table-filters";
import type { ListQuery } from "@/lib/services/swantara";
import { exportCsv } from "@/lib/utils";
import { getColumnLabel, getColumnValue } from "@/lib/utils/table";

export type ServerEntityTableProps<T, V, TResponse> = {
  resource: string;
  fetcher: (
    organizationId: number,
    params: ListQuery,
  ) => Promise<
    TResponse & {
      meta?: {
        pagination?: {
          page?: number;
          perPage?: number;
          total?: number;
          totalPages?: number;
        };
      };
    }
  >;
  selectData: (response: TResponse) => T[];
  columns: ColumnDef<T, V>[];
  getRowId: (row: T, index: number) => string;
  searchKeys: string[];
  statusKey?: keyof T;
  statusOptions?: Array<{ value: string; label: string }>;
  searchPlaceholder: string;
  filterLabel?: string;
  allLabel?: string;
  ariaLabel: string;
  pageSize?: number;
  actions?: ReactNode;
  emptyTitle?: string;
  emptyDescription?: string;
  exportFileName?: string;
};

function buildFilterConfig(
  columns: ColumnDef<unknown, unknown>[],
  searchKeys: string[],
  statusKey?: string,
  statusOptions?: Array<{ value: string; label: string }>,
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

export function ServerEntityTable<T, V, TResponse>({
  resource,
  fetcher,
  selectData,
  columns,
  getRowId,
  searchKeys,
  statusKey,
  statusOptions = [],
  searchPlaceholder,
  filterLabel = "",
  ariaLabel,
  pageSize = 10,
  actions,
  emptyTitle,
  emptyDescription,
  exportFileName,
}: ServerEntityTableProps<T, V, TResponse>) {
  const tTables = useTranslations("Tables" as unknown as "Common");
  const tx = tTables as unknown as (
    key: string,
    values?: Record<string, string | number>,
  ) => string;
  const tFilters = useTranslations("Filters" as unknown as "Common");
  const tf = tFilters as unknown as (
    key: string,
    values?: Record<string, string | number>,
  ) => string;
  const { pagination, setPagination, sorting, setSorting } = useDataTableState({
    initialPageSize: pageSize,
  });

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
        (label: string) => tf("filterBy", { label }),
      ),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [columns, searchKeys, statusKey, statusOptions, searchPlaceholder, filterLabel, tTables, tf],
  );

  const filters = useTableFilters(filterConfig);

  const listQuery = tableStateToListQuery({
    pagination,
    sorting,
    globalFilter: filters.globalFilter,
    columnFilters: filters.columnFilters,
    searchKeys,
  });

  const handleSortingChange: OnChangeFn<SortingState> = (updater) => {
    setSorting(typeof updater === "function" ? updater(sorting) : updater);
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  };

  const handleGlobalFilterChange: OnChangeFn<string> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: string) => string)(filters.globalFilter)
        : updater;
    filters.onGlobalFilterChange(next);
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  };

  const handleColumnFiltersChange: OnChangeFn<ColumnFiltersState> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: ColumnFiltersState) => ColumnFiltersState)(filters.columnFilters)
        : updater;
    filters.onColumnFiltersChange(next);
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  };

  function handleResetFilters() {
    filters.resetFilters();
    setPagination((prev) => ({ ...prev, pageIndex: 0 }));
  }

  const query = useOrgListQuery<TResponse, ListQuery>(resource, fetcher, listQuery);

  const data = query.data ? selectData(query.data) : [];
  const meta = (
    query.data as
      | {
          meta?: {
            pagination?: {
              page?: number;
              perPage?: number;
              total?: number;
              totalPages?: number;
            };
          };
        }
      | undefined
  )?.meta?.pagination;
  const rowCount = meta?.total ?? data.length;
  const pageCount = meta?.totalPages ?? Math.ceil(rowCount / pagination.pageSize);

  useEffect(() => {
    if (meta?.page == null && meta?.perPage == null) return;
    setPagination((prev) => {
      const pageIndex = meta?.page != null ? meta.page - 1 : prev.pageIndex;
      const pageSize = meta?.perPage ?? prev.pageSize;
      if (pageIndex === prev.pageIndex && pageSize === prev.pageSize) return prev;
      return { ...prev, pageIndex, pageSize };
    });
  }, [meta?.page, meta?.perPage, setPagination]);

  const status: DataTableStatus | undefined = query.isLoading
    ? { type: "loading" }
    : query.isError
      ? {
          type: "error",
          message: (query.error as Error).message,
          onRetry: () => void query.refetch(),
        }
      : undefined;

  return (
    <DataTable
      data={data}
      columns={columns}
      getRowId={getRowId}
      enableSorting
      enablePagination
      enableColumnFilters={filterConfig.some((c) => c.id !== "search")}
      enableGlobalFilter
      manualPagination
      manualSorting
      manualFiltering
      rowCount={rowCount}
      pageCount={pageCount}
      pagination={pagination}
      sorting={sorting}
      globalFilter={filters.globalFilter}
      columnFilters={filters.columnFilters}
      onPaginationChange={setPagination}
      onSortingChange={handleSortingChange}
      onGlobalFilterChange={handleGlobalFilterChange}
      onColumnFiltersChange={handleColumnFiltersChange}
      ariaLabel={ariaLabel}
      status={
        status ??
        (data.length === 0
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
          onGlobalFilterChange={handleGlobalFilterChange}
          onReset={handleResetFilters}
          className="w-full"
          actions={
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const headers = columns.map((column) => getColumnLabel(column));
                  const rows = data.map((row) =>
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
