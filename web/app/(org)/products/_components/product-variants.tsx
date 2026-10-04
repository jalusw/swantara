"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { ActiveBadge } from "@/components/active-badge";
import { Badge } from "@/components/badge";
import type { ItemVariant } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function ItemVariants({ variants }: { variants: ItemVariant[] }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const columns: ColumnDef<ItemVariant>[] = [
    {
      accessorKey: "sku",
      header: () => "SKU",
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.sku ?? "—"}</span>,
    },
    {
      accessorKey: "barcode",
      header: () => t("colBarcode"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.barcode ?? "—"}</span>
      ),
    },
    {
      id: "attributes",
      header: () => t("colAttributes"),
      cell: ({ row }) => {
        const entries = row.original.attributeJson
          ? Object.entries(row.original.attributeJson)
          : [];
        if (entries.length === 0) {
          return <span className="text-muted-foreground">—</span>;
        }
        return (
          <div className="flex flex-wrap gap-1">
            {entries.map(([name, value]) => (
              <Badge key={name} variant="secondary">
                {name}: {String(value)}
              </Badge>
            ))}
          </div>
        );
      },
    },
    {
      accessorKey: "extraCost",
      header: () => t("colExtraCost"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.extraCost)}
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
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={variants}
      getRowId={(row) => String(row.id)}
      searchKeys={["sku", "barcode"]}
      statusKey="active"
      statusOptions={[
        { value: "true", label: t("active") },
        { value: "false", label: t("inactive") },
      ]}
      searchPlaceholder={t("searchVariants")}
      filterLabel={t("colStatus")}
      allLabel={t("filterAll")}
      ariaLabel="SKU"
      emptyTitle={t("variantsEmpty")}
    />
  );
}
