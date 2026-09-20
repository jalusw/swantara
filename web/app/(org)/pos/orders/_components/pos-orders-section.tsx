"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { PosOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { posOrderStateLabel, posOrderStateTone } from "../../_components/pos-utils";

export function PosOrdersSection() {
  const ordersQuery = useOrgListQuery<{ orders: PosOrder[] }, Record<string, never>>(
    "posOrders",
    (organizationId) => getSwantaraService().posOrders.list(organizationId),
  );

  const orders = ordersQuery.data?.orders ?? [];

  const columns: ColumnDef<PosOrder>[] = [
    {
      accessorKey: "name",
      header: "Order",
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
      header: "Session",
      cell: ({ row }) => (
        <a
          href={`/pos/sessions/${row.original.sessionId}`}
          className="text-sm text-muted-foreground hover:underline"
        >
          Session-{row.original.sessionId}
        </a>
      ),
    },
    {
      accessorKey: "amountTotal",
      header: "Total",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatMoney(row.original.amountTotal)}</span>
      ),
    },
    {
      accessorKey: "amountTax",
      header: "Tax",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatMoney(row.original.amountTax)}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <Badge variant="outline" className={posOrderStateTone(row.original.state)}>
          {posOrderStateLabel(row.original.state)}
        </Badge>
      ),
    },
    {
      accessorKey: "invoiceId",
      header: "Invoice",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.invoiceId ? `#${row.original.invoiceId}` : "—"}
        </span>
      ),
    },
    {
      accessorKey: "orderTime",
      header: "Time",
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
          { value: "done", label: "Done" },
          { value: "refunded", label: "Refunded" },
        ]}
        searchPlaceholder={"Search orders..."}
        filterLabel={"State"}
        allLabel={"All orders"}
        ariaLabel={"POS Orders"}
        emptyTitle={"No POS orders"}
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
