"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Payment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { paymentStateTone } from "./payment-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function PaymentsSection() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const paymentsQuery = useOrgListQuery<{ payments: Payment[] }, Record<string, never>>(
    "payments",
    (organizationId) => getSwantaraService().payments.list(organizationId),
  );

  const payments = paymentsQuery.data?.payments ?? [];

  const columns: ColumnDef<Payment>[] = [
    {
      accessorKey: "name",
      header: () => t("colNumber"),
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
      header: () => t("colContact"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{row.original.contactId ?? "—"}</span>
      ),
    },
    {
      accessorKey: "type",
      header: () => t("colDirection"),
      cell: ({ row }) => (
        <Badge variant="secondary">
          {(t as unknown as (k: string) => string)(`paymentDirection_${row.original.type}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "date",
      header: () => t("colDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.date)}</span>
      ),
    },
    {
      accessorKey: "amount",
      header: () => t("colAmount"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatMoney(row.original.amount)}</span>,
    },
    {
      accessorKey: "state",
      header: () => t("colStatus"),
      cell: ({ row }) => (
        <Badge variant="outline" className={paymentStateTone(row.original.state)}>
          {(t as unknown as (k: string) => string)(`paymentState_${row.original.state}`)}
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
          { value: "draft", label: (t as unknown as (k: string) => string)("paymentState_draft") },
          {
            value: "posted",
            label: (t as unknown as (k: string) => string)("paymentState_posted"),
          },
          {
            value: "reconciled",
            label: (t as unknown as (k: string) => string)("paymentState_reconciled"),
          },
          {
            value: "cancelled",
            label: (t as unknown as (k: string) => string)("paymentState_cancelled"),
          },
        ]}
        searchPlaceholder={t("searchPayments")}
        filterLabel={t("colStatus")}
        allLabel={t("filterAll")}
        ariaLabel={t("paymentsTitle")}
        emptyTitle={t("paymentsEmpty")}
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
