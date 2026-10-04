"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { StockCount } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { nameColumn } from "@/lib/utils/table-columns";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function toRow(count: StockCount) {
  return {
    id: String(count.id),
    name: count.name ?? `IC-${count.id}`,
    state: count.state,
    countDate: count.countDate,
  };
}

export function CountsSection() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Stock");
  const countsQuery = useOrgListQuery<{ counts: StockCount[] }, Record<string, never>>(
    "inventoryCounts",
    (organizationId) => getSwantaraService().inventory.inventoryCounts(organizationId),
  );

  const rows = (countsQuery.data?.counts ?? []).map(toRow);

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    nameColumn<ReturnType<typeof toRow>>({
      basePath: "stock/counts",
      header: t("colCount"),
    }),
    {
      accessorKey: "state",
      header: () => t("colStatus"),
      cell: ({ row }) => (
        <StateBadge
          value={row.original.state}
          statuses={{
            posted: {
              label: (t as unknown as (k: string) => string)("countState_posted"),
              tone: "success",
            },
            draft: {
              label: (t as unknown as (k: string) => string)("countState_draft"),
              tone: "neutral",
            },
          }}
        />
      ),
    },
    {
      accessorKey: "countDate",
      header: () => t("colCountDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.countDate ? formatDate(row.original.countDate, { nullFallback: "—" }) : "—"}
        </span>
      ),
    },
  ];

  const isLoading = countsQuery.isLoading;
  const error = countsQuery.isError ? countsQuery.error : null;

  return (
    <InteractiveEntityTable
      columns={columns}
      data={rows}
      getRowId={(row) => row.id}
      searchKeys={["name"]}
      searchPlaceholder={t("searchCounts")}
      ariaLabel={t("allCounts")}
      emptyTitle={t("countsEmpty")}
      status={
        isLoading
          ? { type: "loading" }
          : error
            ? {
                type: "error",
                message: error.message,
                onRetry: () => void countsQuery.refetch(),
              }
            : undefined
      }
    />
  );
}
