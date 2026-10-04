"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { PosOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { posOrderStateTone } from "../../_components/pos-utils";

export function PosOrdersSection() {
  const t = useTranslations("Pos");
  const orderState = (state: string) =>
    (t as unknown as (k: string) => string)(`orderState_${state}`);
  const ordersQuery = useOrgListQuery<{ orders: PosOrder[] }, Record<string, never>>(
    "posOrders",
    (organizationId) => getSwantaraService().posOrders.list(organizationId),
  );

  const orders = ordersQuery.data?.orders ?? [];

  const columns: ColumnDef<PosOrder>[] = [
    {
      accessorKey: "name",
      header: t("colOrder"),
      cell: ({ row }) => (
        <a
          href={`/pos/orders/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `POS-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "sessionId",
      header: t("colSession"),
      cell: ({ row }) => (
        <a
          href={`/pos/sessions/${row.original.sessionId}`}
          className="text-sm text-muted-foreground hover:underline"
        >
          {t("sessionFallback", { id: row.original.sessionId })}
        </a>
      ),
    },
    {
      accessorKey: "amountTotal",
      header: t("colTotal"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatMoney(row.original.amountTotal)}</span>
      ),
    },
    {
      accessorKey: "amountTax",
      header: t("colTax"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatMoney(row.original.amountTax)}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => (
        <Badge variant="outline" className={posOrderStateTone(row.original.state)}>
          {orderState(row.original.state)}
        </Badge>
      ),
    },
    {
      accessorKey: "invoiceId",
      header: t("colInvoice"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.invoiceId ? `#${row.original.invoiceId}` : "—"}
        </span>
      ),
    },
    {
      accessorKey: "orderTime",
      header: t("colTime"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.orderTime ? formatDate(row.original.orderTime) : "—"}
        </span>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={orders}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "done", label: orderState("done") },
          { value: "refunded", label: orderState("refunded") },
        ]}
        searchPlaceholder={t("ordersSearchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allOrders")}
        ariaLabel={t("ordersTitle")}
        emptyTitle={t("ordersEmpty")}
        status={
          ordersQuery.isLoading
            ? { type: "loading" }
            : ordersQuery.isError
              ? {
                  type: "error",
                  message: ordersQuery.error.message,
                  onRetry: () => void ordersQuery.refetch(),
                }
              : undefined
        }
      />
    </div>
  );
}
