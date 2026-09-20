"use client";

import type { ColumnDef } from "@tanstack/react-table";
import type { ReactNode } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { ActiveBadge } from "@/components/active-badge";
import { Badge } from "@/components/badge";
import { formatNumber } from "@/lib/utils";
import { bomLineTotalQty, type StubBom, type StubItem } from "../../_components/products-data";

export function BomsTable({
  recipes,
  templates,
  renderActions,
  status,
}: {
  recipes: StubBom[];
  templates: StubItem[];
  renderActions?: (recipe: StubBom) => ReactNode;
  status?: { type: "loading" } | { type: "error"; message: string; onRetry: () => void };
}) {
  const columns: ColumnDef<StubBom>[] = [
    {
      accessorKey: "code",
      header: "Reference",
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.code ?? "—"}</span>,
    },
    {
      accessorKey: "itemId",
      header: "Item",
      cell: ({ row }) => (
        <span className="">
          {templates.find((template) => template.id === row.original.itemId)?.name ?? "—"}
        </span>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => <Badge variant="secondary">{String(row.original.type)}</Badge>,
    },
    {
      accessorKey: "version",
      header: "Version",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">v{row.original.version}</span>
      ),
    },
    {
      accessorKey: "qty",
      header: "Quantity",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{formatNumber(row.original.qty)}</span>
      ),
    },
    {
      id: "lines",
      header: "Lines",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground" title={"With scrap"}>
          {formatNumber(row.original.lines.reduce((sum, line) => sum + bomLineTotalQty(line), 0))}
        </span>
      ),
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
    ...(renderActions
      ? [
          {
            id: "actions",
            header: "",
            cell: ({ row }: { row: { original: StubBom } }) => renderActions(row.original),
          } satisfies ColumnDef<StubBom>,
        ]
      : []),
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={recipes}
      getRowId={(row) => row.id}
      searchKeys={["code"]}
      statusKey="active"
      statusOptions={[
        { value: "true", label: "Active" },
        { value: "false", label: "Inactive" },
      ]}
      searchPlaceholder={"Search bills of materials…"}
      filterLabel={"Filter by status"}
      allLabel={"All statuses"}
      ariaLabel={"All bills of materials"}
      emptyTitle={"No bills of materials"}
      status={status}
    />
  );
}
