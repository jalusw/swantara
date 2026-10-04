"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { WarehouseTransfer } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { makeStateVariant, nameColumn, stateColumn } from "@/lib/utils/table-columns";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function toRow(warehouseTransfer: WarehouseTransfer) {
  return {
    id: String(warehouseTransfer.id),
    name: warehouseTransfer.name ?? `TO-${warehouseTransfer.id}`,
    state: warehouseTransfer.state,
    scheduledDate: warehouseTransfer.scheduledDate,
    srcWarehouseId: warehouseTransfer.srcWarehouseId,
    dstWarehouseId: warehouseTransfer.dstWarehouseId,
  };
}

export function TransfersSection() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Stock");
  const transfersQuery = useOrgListQuery<
    { warehouseTransfers: WarehouseTransfer[] },
    Record<string, never>
  >("warehouseTransfers", (organizationId) =>
    getSwantaraService().inventory.transferOrders(organizationId),
  );

  const rows = (transfersQuery.data?.warehouseTransfers ?? []).map(toRow);

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    nameColumn<ReturnType<typeof toRow>>({
      basePath: "stock/warehouse-transfers",
      header: t("colTransfer"),
    }),
    stateColumn<ReturnType<typeof toRow>>({
      header: t("colStatus"),
      label: (state) => (t as unknown as (k: string) => string)(`transferState_${String(state)}`),
      variant: makeStateVariant(["received"]),
    }),
    {
      accessorKey: "scheduledDate",
      header: () => t("colScheduled"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.scheduledDate
            ? formatDate(row.original.scheduledDate, { nullFallback: "—" })
            : "—"}
        </span>
      ),
    },
  ];

  const isLoading = transfersQuery.isLoading;
  const error = transfersQuery.isError ? transfersQuery.error : null;

  return (
    <InteractiveEntityTable
      columns={columns}
      data={rows}
      getRowId={(row) => row.id}
      searchKeys={["name"]}
      searchPlaceholder={t("searchTransfers")}
      ariaLabel={t("allTransfers")}
      emptyTitle={t("transfersEmpty")}
      status={
        isLoading
          ? { type: "loading" }
          : error
            ? {
                type: "error",
                message: error.message,
                onRetry: () => void transfersQuery.refetch(),
              }
            : undefined
      }
    />
  );
}
