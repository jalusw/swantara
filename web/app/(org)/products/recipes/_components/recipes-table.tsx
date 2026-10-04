"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import type { ReactNode } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { ActiveBadge } from "@/components/active-badge";
import { Badge } from "@/components/badge";
import { formatNumber } from "@/lib/utils";
import { bomLineTotalQty, type StubBom, type StubItem } from "../../_components/products-data";

type TFn = (key: string, values?: Record<string, string | number>) => string;

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
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const columns: ColumnDef<StubBom>[] = [
    {
      accessorKey: "code",
      header: () => t("colReference"),
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.code ?? "—"}</span>,
    },
    {
      accessorKey: "itemId",
      header: () => t("fieldItem"),
      cell: ({ row }) => (
        <span className="">
          {templates.find((template) => template.id === row.original.itemId)?.name ?? "—"}
        </span>
      ),
    },
    {
      accessorKey: "type",
      header: () => t("fieldType"),
      cell: ({ row }) => (
        <Badge variant="secondary">
          {(t as unknown as (k: string) => string)(`recipeType_${row.original.type}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "version",
      header: () => t("colVersion"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">v{row.original.version}</span>
      ),
    },
    {
      accessorKey: "qty",
      header: () => t("colQuantity"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{formatNumber(row.original.qty)}</span>
      ),
    },
    {
      id: "lines",
      header: () => t("lines"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground" title={t("withScrap")}>
          {formatNumber(row.original.lines.reduce((sum, line) => sum + bomLineTotalQty(line), 0))}
        </span>
      ),
    },
    {
      accessorKey: "active",
      header: () => t("colStatus"),
      cell: ({ row }) => (
        <ActiveBadge active={row.original.active}>
          {row.original.active ? t("active") : t("inactive")}
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
        { value: "true", label: t("active") },
        { value: "false", label: t("inactive") },
      ]}
      searchPlaceholder={t("searchRecipes")}
      filterLabel={t("filterByStatus")}
      allLabel={t("filterAllStatus")}
      ariaLabel={t("allRecipes")}
      emptyTitle={t("recipesEmpty")}
      status={status}
    />
  );
}
