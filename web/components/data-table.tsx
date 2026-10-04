"use client";

import {
  type ColumnDef,
  type ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  type OnChangeFn,
  type PaginationState,
  type SortingState,
  useReactTable,
  type VisibilityState,
} from "@tanstack/react-table";
import {
  ArrowDownIcon,
  ArrowUpIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsUpDownIcon,
  InboxIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { useMessages, useTranslations } from "next-intl";
import type { ReactNode } from "react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { getColumnLabel } from "@/lib/utils/table";
import { Button } from "./button";
import { Checkbox } from "./checkbox";
import { EmptyState } from "./empty-state";
import { RouteLoading } from "./route-loading";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./table";

export type DataTableStatus =
  | { type: "loading"; rows?: number }
  | { type: "error"; message?: string; onRetry?: () => void }
  | { type: "empty"; title?: string; description?: string }
  | { type: "none" };

export type DataTableColumnMeta = {
  align?: "left" | "right" | "center";
  className?: string;
};

export type DataTableLabels = {
  selectAll?: string;
  selectRow?: string;
  sortBy?: string;
  errorMessage?: string;
  retry?: string;
  emptyTitle?: string;
  emptyDescription?: string;
  rowsSelected?: string;
  rowsPerPage?: string;
  page?: string;
  of?: string;
  pagePrevious?: string;
  pageNext?: string;
  range?: string;
};

export type DataTableProps<TData, TValue> = {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  getRowId: (row: TData, index: number) => string;
  enableRowSelection?: boolean;
  enableSorting?: boolean;
  enablePagination?: boolean;
  enableColumnFilters?: boolean;
  enableGlobalFilter?: boolean;
  enableHiding?: boolean;
  manualPagination?: boolean;
  manualSorting?: boolean;
  manualFiltering?: boolean;
  rowCount?: number;
  pageCount?: number;
  initialPageSize?: number;
  pageSizeOptions?: number[];
  initialSorting?: SortingState;
  initialColumnFilters?: ColumnFiltersState;
  initialGlobalFilter?: string;
  initialColumnVisibility?: VisibilityState;
  sorting?: SortingState;
  pagination?: PaginationState;
  columnFilters?: ColumnFiltersState;
  globalFilter?: string;
  columnVisibility?: VisibilityState;
  onSortingChange?: OnChangeFn<SortingState>;
  onPaginationChange?: OnChangeFn<PaginationState>;
  onColumnFiltersChange?: OnChangeFn<ColumnFiltersState>;
  onGlobalFilterChange?: OnChangeFn<string>;
  onColumnVisibilityChange?: OnChangeFn<VisibilityState>;
  onSelectionChange?: (selectedKeys: string[]) => void;
  status?: DataTableStatus;
  toolbar?: ReactNode;
  ariaLabel?: string;
  className?: string;
  labels?: DataTableLabels;
};

function useDefaultLabels(): Required<DataTableLabels> {
  const raw = useMessages() as unknown as { Tables?: Record<string, string> };
  const messages = raw.Tables ?? {};
  return {
    selectAll: messages.selectAll ?? "",
    selectRow: messages.selectRow ?? "",
    sortBy: messages.sortBy ?? "",
    errorMessage: messages.errorMessage ?? "",
    retry: messages.retry ?? "",
    emptyTitle: messages.emptyTitle ?? "",
    emptyDescription: messages.emptyDescription ?? "",
    rowsSelected: messages.rowsSelected ?? "",
    rowsPerPage: messages.rowsPerPage ?? "",
    page: messages.page ?? "",
    of: messages.of ?? "",
    pagePrevious: messages.pagePrevious ?? "",
    pageNext: messages.pageNext ?? "",
    range: messages.range ?? "",
  };
}

const alignClass: Record<NonNullable<DataTableColumnMeta["align"]>, string> = {
  left: "text-left",
  right: "text-right",
  center: "text-center",
};

function format(template: string, vars: Record<string, string | number>): string {
  let result = template;
  for (const [key, value] of Object.entries(vars)) {
    result = result.replaceAll(`{${key}}`, String(value));
  }
  return result;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  getRowId,
  enableRowSelection = false,
  enableSorting = false,
  enablePagination = false,
  enableColumnFilters = false,
  enableGlobalFilter = false,
  enableHiding = false,
  manualPagination = false,
  manualSorting = false,
  manualFiltering = false,
  rowCount,
  pageCount,
  initialPageSize = 10,
  pageSizeOptions = [10, 25, 50],
  initialSorting,
  initialColumnFilters,
  initialGlobalFilter,
  initialColumnVisibility,
  sorting: controlledSorting,
  pagination: controlledPagination,
  columnFilters: controlledColumnFilters,
  globalFilter: controlledGlobalFilter,
  columnVisibility: controlledVisibility,
  onSortingChange,
  onPaginationChange,
  onColumnFiltersChange,
  onGlobalFilterChange,
  onColumnVisibilityChange,
  onSelectionChange,
  status = { type: "none" },
  toolbar,
  ariaLabel,
  className,
  labels,
}: DataTableProps<TData, TValue>) {
  const tAria = useTranslations("Tables" as unknown as "Common");
  const txAria = tAria as unknown as (key: string) => string;
  const resolvedAriaLabel = ariaLabel ?? txAria("ariaLabel");
  const translatedDefaults = useDefaultLabels();
  const t = { ...translatedDefaults, ...labels };
  const [internalSorting, setInternalSorting] = useState<SortingState>(initialSorting ?? []);
  const [internalPagination, setInternalPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: initialPageSize,
  });
  const [internalColumnFilters, setInternalColumnFilters] = useState<ColumnFiltersState>(
    initialColumnFilters ?? [],
  );
  const [internalGlobalFilter, setInternalGlobalFilter] = useState(initialGlobalFilter ?? "");
  const [internalVisibility, setInternalVisibility] = useState<VisibilityState>(
    initialColumnVisibility ?? {},
  );
  const [rowSelection, setRowSelection] = useState<Record<string, boolean>>({});

  const sorting = controlledSorting ?? internalSorting;
  const pagination = controlledPagination ?? internalPagination;
  const columnFilters = controlledColumnFilters ?? internalColumnFilters;
  const globalFilter = controlledGlobalFilter ?? internalGlobalFilter;
  const columnVisibility = controlledVisibility ?? internalVisibility;

  const handleSortingChange: OnChangeFn<SortingState> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: SortingState) => SortingState)(sorting)
        : updater;
    if (controlledSorting === undefined) {
      setInternalSorting(next);
    }
    onSortingChange?.(next);
  };

  const handlePaginationChange: OnChangeFn<PaginationState> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: PaginationState) => PaginationState)(pagination)
        : updater;
    if (controlledPagination === undefined) {
      setInternalPagination(next);
    }
    onPaginationChange?.(next);
  };

  const handleColumnFiltersChange: OnChangeFn<ColumnFiltersState> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: ColumnFiltersState) => ColumnFiltersState)(columnFilters)
        : updater;
    if (controlledColumnFilters === undefined) {
      setInternalColumnFilters(next);
    }
    onColumnFiltersChange?.(next);
  };

  const handleGlobalFilterChange: OnChangeFn<string> = (updater) => {
    const next =
      typeof updater === "function" ? (updater as (old: string) => string)(globalFilter) : updater;
    if (controlledGlobalFilter === undefined) {
      setInternalGlobalFilter(next);
    }
    onGlobalFilterChange?.(next);
  };

  const handleVisibilityChange: OnChangeFn<VisibilityState> = (updater) => {
    const next =
      typeof updater === "function"
        ? (updater as (old: VisibilityState) => VisibilityState)(columnVisibility)
        : updater;
    if (controlledVisibility === undefined) {
      setInternalVisibility(next);
    }
    onColumnVisibilityChange?.(next);
  };

  const enableClientSorting = enableSorting && !manualSorting;
  const enableClientPagination = enablePagination && !manualPagination;
  const enableClientFiltering = (enableColumnFilters || enableGlobalFilter) && !manualFiltering;

  const table = useReactTable({
    data,
    columns,
    getRowId,
    state: {
      sorting,
      pagination,
      rowSelection,
      columnFilters,
      globalFilter,
      columnVisibility,
    },
    onSortingChange: handleSortingChange,
    onPaginationChange: handlePaginationChange,
    onRowSelectionChange: (updater) => {
      const next = typeof updater === "function" ? updater(rowSelection) : updater;
      setRowSelection(next);
      onSelectionChange?.(Object.keys(next));
    },
    onColumnFiltersChange: handleColumnFiltersChange,
    onGlobalFilterChange: handleGlobalFilterChange,
    onColumnVisibilityChange: handleVisibilityChange,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: enableClientSorting ? getSortedRowModel() : undefined,
    getPaginationRowModel: enableClientPagination ? getPaginationRowModel() : undefined,
    getFilteredRowModel: enableClientFiltering ? getFilteredRowModel() : undefined,
    getFacetedRowModel: enableClientFiltering ? getFacetedRowModel() : undefined,
    getFacetedUniqueValues: enableClientFiltering ? getFacetedUniqueValues() : undefined,
    manualPagination,
    manualSorting,
    manualFiltering,
    enableRowSelection,
    enableSorting,
    enableColumnFilters,
    enableGlobalFilter,
    enableHiding,
    rowCount,
    pageCount,
    globalFilterFn: "includesString",
  });

  const columnSpan = table.getVisibleLeafColumns().length + (enableRowSelection ? 1 : 0);
  const selectedCount = table.getIsSomeRowsSelected() ? Object.keys(rowSelection).length : 0;

  const totalRows = manualPagination
    ? (rowCount ?? data.length)
    : table.getFilteredRowModel().rows.length;
  const filteredRowCount = table.getFilteredRowModel().rows.length;
  const hasActiveFilter = columnFilters.length > 0 || (globalFilter?.length ?? 0) > 0;

  let rangeLabel: string | null = null;
  if (enablePagination && totalRows > 0) {
    const { pageIndex, pageSize } = table.getState().pagination;
    const from = pageIndex * pageSize + 1;
    const to = Math.min(from + pageSize - 1, totalRows);
    rangeLabel = format(t.range, {
      from,
      to,
      total: totalRows,
    });
  }

  const isEmpty =
    status.type === "empty" ||
    table.getRowModel().rows.length === 0 ||
    (filteredRowCount === 0 && hasActiveFilter && status.type === "none");

  return (
    <div data-slot="data-table" className={cn("space-y-2", className)}>
      {toolbar ? <div className="flex items-center gap-2">{toolbar}</div> : null}

      <div className="overflow-hidden rounded-xl border border-border">
        <Table aria-label={resolvedAriaLabel}>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {enableRowSelection ? (
                  <TableHead className="w-10">
                    <Checkbox
                      aria-label={t.selectAll}
                      checked={table.getIsAllPageRowsSelected()}
                      indeterminate={table.getIsSomePageRowsSelected()}
                      onCheckedChange={(checked) => table.toggleAllPageRowsSelected(checked)}
                    />
                  </TableHead>
                ) : null}
                {headerGroup.headers.map((header) => {
                  if (header.isPlaceholder) {
                    return null;
                  }
                  const meta = header.column.columnDef.meta as DataTableColumnMeta | undefined;
                  const sort = header.column.getIsSorted();
                  const sortIconMap: Record<string, typeof ArrowUpIcon> = {
                    asc: ArrowUpIcon,
                    desc: ArrowDownIcon,
                  };
                  const SortIcon = sortIconMap[sort as string] ?? ChevronsUpDownIcon;

                  return (
                    <TableHead
                      key={header.id}
                      scope="col"
                      className={cn(alignClass[meta?.align ?? "left"], meta?.className)}
                      aria-sort={
                        sort === "asc" ? "ascending" : sort === "desc" ? "descending" : undefined
                      }
                    >
                      {header.column.getCanSort() && enableSorting ? (
                        <Button
                          variant="ghost"
                          size="xs"
                          className="-ml-1 px-1.5 "
                          aria-label={format(t.sortBy, {
                            label: getColumnLabel(header.column.columnDef),
                          })}
                          onClick={header.column.getToggleSortingHandler()}
                        >
                          {flexRender(header.column.columnDef.header, header.getContext())}
                          <SortIcon className="size-3.5 text-muted-foreground" aria-hidden />
                        </Button>
                      ) : (
                        flexRender(header.column.columnDef.header, header.getContext())
                      )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>

          {status.type === "loading" ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={columnSpan} className="p-0">
                  <RouteLoading variant="table" rows={status.rows ?? 3} />
                </TableCell>
              </TableRow>
            </TableBody>
          ) : null}

          {status.type === "error" ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={columnSpan} className="py-8">
                  <div className="flex flex-col items-center gap-2 text-sm">
                    <TriangleAlertIcon className="size-5 text-destructive" aria-hidden />
                    <span className="text-muted-foreground">
                      {status.message ?? t.errorMessage}
                    </span>
                    {status.onRetry ? (
                      <Button variant="outline" size="sm" onClick={status.onRetry}>
                        {t.retry}
                      </Button>
                    ) : null}
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          ) : null}

          {isEmpty && status.type !== "loading" && status.type !== "error" ? (
            <TableBody>
              <TableRow>
                <TableCell colSpan={columnSpan} className="p-0">
                  <EmptyState
                    icon={InboxIcon}
                    title={status.type === "empty" ? (status.title ?? t.emptyTitle) : t.emptyTitle}
                    description={status.type === "empty" ? status.description : t.emptyDescription}
                  />
                </TableCell>
              </TableRow>
            </TableBody>
          ) : null}

          {status.type === "none" && table.getRowModel().rows.length > 0 && !isEmpty ? (
            <TableBody>
              {table.getRowModel().rows.map((row) => {
                return (
                  <TableRow key={row.id} data-state={row.getIsSelected() ? "selected" : undefined}>
                    {enableRowSelection ? (
                      <TableCell>
                        <Checkbox
                          aria-label={format(t.selectRow, { id: row.id })}
                          checked={row.getIsSelected()}
                          onCheckedChange={row.getToggleSelectedHandler()}
                        />
                      </TableCell>
                    ) : null}
                    {row.getVisibleCells().map((cell) => {
                      const cellMeta = cell.column.columnDef.meta as
                        | DataTableColumnMeta
                        | undefined;
                      return (
                        <TableCell
                          key={cell.id}
                          className={cn(alignClass[cellMeta?.align ?? "left"], cellMeta?.className)}
                        >
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                );
              })}
            </TableBody>
          ) : null}
        </Table>
      </div>

      {enablePagination && status.type !== "loading" ? (
        <footer className="flex flex-col items-center gap-3 pt-1 sm:flex-row sm:justify-between">
          <p
            className="text-sm text-muted-foreground"
            aria-live={selectedCount > 0 ? "polite" : undefined}
          >
            {selectedCount > 0 ? format(t.rowsSelected, { count: selectedCount }) : rangeLabel}
          </p>
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2">
              <span className="sr-only">{t.rowsPerPage}</span>
              <Select
                value={String(table.getState().pagination.pageSize)}
                onValueChange={(value) => {
                  table.setPageSize(Number(value));
                }}
              >
                <SelectTrigger size="sm" aria-label={t.rowsPerPage}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {pageSizeOptions.map((size) => (
                    <SelectItem key={size} value={String(size)}>
                      {size} / {t.page}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="icon-sm"
                aria-label={t.pagePrevious}
                disabled={!table.getCanPreviousPage()}
                onClick={() => table.previousPage()}
              >
                <ChevronLeftIcon aria-hidden />
              </Button>
              <span className="min-w-16 px-1 text-center text-sm text-muted-foreground">
                {t.page} {table.getState().pagination.pageIndex + 1} {t.of}{" "}
                {manualPagination && pageCount != null
                  ? pageCount
                  : Math.max(table.getPageCount(), 1)}
              </span>
              <Button
                variant="outline"
                size="icon-sm"
                aria-label={t.pageNext}
                disabled={!table.getCanNextPage()}
                onClick={() => table.nextPage()}
              >
                <ChevronRightIcon aria-hidden />
              </Button>
            </div>
          </div>
        </footer>
      ) : null}
    </div>
  );
}
