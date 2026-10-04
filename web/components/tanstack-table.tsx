"use client";

import { type DataTableProps, DataTable as UiDataTable } from "@/components/data-table";

export type {
  DataTableColumnMeta,
  DataTableProps,
  DataTableStatus,
} from "@/components/data-table";

export function DataTable<TData, TValue>(props: Omit<DataTableProps<TData, TValue>, "labels">) {
  return <UiDataTable data-slot="data-table" {...(props as DataTableProps<TData, TValue>)} />;
}
