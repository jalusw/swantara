import type { ColumnDef } from "@tanstack/react-table";

export function getColumnLabel<T, V>(column: ColumnDef<T, V>): string {
  if (typeof column.header === "string") {
    return column.header;
  }
  if ("accessorKey" in column && column.accessorKey) {
    return String(column.accessorKey);
  }
  return column.id ?? "";
}

export function resolvePath(row: unknown, path: string): unknown {
  return path.split(".").reduce<unknown>((value, key) => {
    if (value == null) {
      return undefined;
    }
    return (value as Record<string, unknown>)[key];
  }, row);
}

export function getColumnValue<T, V>(column: ColumnDef<T, V>, row: T): unknown {
  if ("accessorFn" in column && column.accessorFn) {
    return column.accessorFn(row, 0);
  }
  if ("accessorKey" in column && column.accessorKey) {
    return resolvePath(row, String(column.accessorKey));
  }
  return undefined;
}
