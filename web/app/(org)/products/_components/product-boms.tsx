"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { ActiveBadge } from "@/components/active-badge";
import type { Recipe } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function ProductBoms({
  recipes,
  status,
}: {
  recipes: Recipe[];
  status?: { type: "loading" } | { type: "error"; message: string; onRetry: () => void };
}) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const columns: ColumnDef<Recipe>[] = [
    {
      accessorKey: "code",
      header: () => t("colReference"),
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.code ?? "—"}</span>,
    },
    {
      accessorKey: "type",
      header: () => t("fieldType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {(t as unknown as (k: string) => string)(`recipeType_${row.original.type}`)}
        </span>
      ),
    },
    {
      accessorKey: "qty",
      header: () => t("colQuantity"),
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.qty)}</span>,
    },
    {
      accessorKey: "version",
      header: () => t("colVersion"),
      cell: ({ row }) => <span className="tabular-nums">v{row.original.version}</span>,
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
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={recipes}
      getRowId={(row) => String(row.id)}
      searchKeys={["code"]}
      searchPlaceholder={t("searchRecipes")}
      filterLabel={t("filterByStatus")}
      allLabel={t("filterAllStatus")}
      ariaLabel={t("allRecipes")}
      statusOptions={[
        { value: "true", label: t("active") },
        { value: "false", label: t("inactive") },
      ]}
      status={status}
      emptyTitle={t("recipesEmpty")}
    />
  );
}
