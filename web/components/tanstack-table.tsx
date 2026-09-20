"use client";

import { type DataTableProps, DataTable as UiDataTable } from "@/components/data-table";

export type {
  DataTableColumnMeta,
  DataTableProps,
  DataTableStatus,
} from "@/components/data-table";

export function DataTable<TData, TValue>(props: Omit<DataTableProps<TData, TValue>, "labels">) {
  return (
    <UiDataTable
      data-slot="data-table"
      {...(props as DataTableProps<TData, TValue>)}
      labels={{
        selectAll: "Select all rows",
        selectRow: "Select row {id}",
        sortBy: "Sort by {label}",
        errorMessage: "Could not load rows.",
        retry: "Retry",
        emptyTitle: "No records found",
        emptyDescription: "No records match your filters.",
        rowsSelected: "{count} selected",
        rowsPerPage: "Rows per page",
        page: "Page",
        of: "of",
        pagePrevious: "Previous page",
        pageNext: "Next page",
        range: "{from}–{to} of {total}",
      }}
    />
  );
}
