"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Payment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { paymentDirectionLabel, paymentStateLabel, paymentStateTone } from "./payment-utils";

export function PaymentsSection() {
  const paymentsQuery = useOrgListQuery<{ payments: Payment[] }, Record<string, never>>(
    "payments",
    (organizationId) => getSwantaraService().payments.list(organizationId),
  );

  const payments = paymentsQuery.data?.payments ?? [];

  const columns: ColumnDef<Payment>[] = [
    {
      accessorKey: "name",
      header: "Number",
      cell: ({ row }) => (
        <a
          href={`/accounting/payments/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `PAY-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "contactId",
      header: "Contact",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{row.original.contactId ?? "—"}</span>
      ),
    },
    {
      accessorKey: "type",
      header: "Direction",
      cell: ({ row }) => (
        <Badge variant="secondary">{paymentDirectionLabel(row.original.type)}</Badge>
      ),
    },
    {
      accessorKey: "date",
      header: "Date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.date)}</span>
      ),
    },
    {
      accessorKey: "amount",
      header: "Amount",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatMoney(row.original.amount)}</span>,
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <Badge variant="outline" className={paymentStateTone(row.original.state)}>
          {paymentStateLabel(row.original.state)}
        </Badge>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={payments}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "posted", label: "Posted" },
          { value: "reconciled", label: "Reconciled" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search payments..."}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"Payments"}
        emptyTitle={"No payments recorded."}
        status={
          paymentsQuery.isLoading
            ? { type: "loading" }
            : paymentsQuery.isError
              ? {
                  type: "error",
                  message: paymentsQuery.error.message,
                  onRetry: () => void paymentsQuery.refetch(),
                }
              : undefined
        }
      />
    </div>
  );
}
