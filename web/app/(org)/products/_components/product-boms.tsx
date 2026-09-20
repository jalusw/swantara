"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { ActiveBadge } from "@/components/active-badge";
import type { Recipe } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

export function ProductBoms({
  recipes,
  status,
}: {
  recipes: Recipe[];
  status?: { type: "loading" } | { type: "error"; message: string; onRetry: () => void };
}) {
  const columns: ColumnDef<Recipe>[] = [
    {
      accessorKey: "code",
      header: "Reference",
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.code ?? "—"}</span>,
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{humanizeKey(String(row.original.type))}</span>
      ),
    },
    {
      accessorKey: "qty",
      header: "Quantity",
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.qty)}</span>,
    },
    {
      accessorKey: "version",
      header: "Version",
      cell: ({ row }) => <span className="tabular-nums">v{row.original.version}</span>,
    },
    {
      accessorKey: "active",
      header: "Status",
      cell: ({ row }) => (
        <ActiveBadge active={row.original.active}>
          {row.original.active ? "Active" : "Inactive"}
        </ActiveBadge>
      ),
    },
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={recipes}
      getRowId={(row) => String(row.id)}
      searchKeys={["code"]}
      searchPlaceholder={"Search bills of materials…"}
      filterLabel={"Filter by status"}
      allLabel={"All statuses"}
      ariaLabel={"All bills of materials"}
      statusOptions={[
        { value: "true", label: "Active" },
        { value: "false", label: "Inactive" },
      ]}
      status={status}
      emptyTitle={"No bills of materials"}
    />
  );
}
