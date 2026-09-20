"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { StockCount } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { nameColumn } from "@/lib/utils/table-columns";

function toRow(count: StockCount) {
  return {
    id: String(count.id),
    name: count.name ?? `IC-${count.id}`,
    state: count.state,
    countDate: count.countDate,
  };
}

export function CountsSection() {
  const countsQuery = useOrgListQuery<{ counts: StockCount[] }, Record<string, never>>(
    "inventoryCounts",
    (organizationId) => getSwantaraService().inventory.inventoryCounts(organizationId),
  );

  const rows = (countsQuery.data?.counts ?? []).map(toRow);

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    nameColumn<ReturnType<typeof toRow>>({
      basePath: "stock/counts",
      header: "Count",
    }),
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <StateBadge
          value={row.original.state}
          statuses={{
            posted: { label: row.original.state, tone: "success" },
            draft: { label: row.original.state, tone: "neutral" },
          }}
        />
      ),
    },
    {
      accessorKey: "countDate",
      header: "Count date",
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
      searchPlaceholder={"Search counts…"}
      ariaLabel={"All counts"}
      emptyTitle={"No inventory counts found."}
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
