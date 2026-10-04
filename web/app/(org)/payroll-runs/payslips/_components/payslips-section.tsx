"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Payslip } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

export function PayslipsSection() {
  const t = useTranslations("Payroll");
  const payslipsQuery = useOrgListQuery<{ payslips: Payslip[] }, Record<string, never>>(
    "payslips",
    (organizationId) => getSwantaraService().payslips.list(organizationId),
  );

  const payslips = payslipsQuery.data?.payslips ?? [];

  function stateLabel(state: Payslip["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`payslipState.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<Payslip>[] = [
    {
      accessorKey: "id",
      header: t("tableNumber"),
      cell: ({ row }) => (
        <a
          href={`/payroll-runs/payslips/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          PS-{row.original.id}
        </a>
      ),
    },
    {
      accessorKey: "employeeId",
      header: t("fieldEmployee"),
      cell: ({ row }) => <span className="text-muted-foreground">#{row.original.employeeId}</span>,
    },
    {
      accessorKey: "gross",
      header: t("grossTotal"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatNumber(row.original.gross)}</span>,
    },
    {
      accessorKey: "net",
      header: t("netTotal"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatNumber(row.original.net)}</span>,
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
      cell: ({ row }) => (
        <Badge variant={row.original.state === "posted" ? "default" : "outline"}>
          {stateLabel(row.original.state)}
        </Badge>
      ),
    },
    {
      accessorKey: "moveId",
      header: t("tableJournalEntry"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.entryId != null ? `#${row.original.entryId}` : "—"}
        </span>
      ),
    },
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={payslips}
      getRowId={(row) => String(row.id)}
      searchKeys={["employeeId"]}
      statusKey="state"
      statusOptions={[
        { value: "draft", label: t("payslipStateDraft") },
        { value: "posted", label: t("payslipStatePosted") },
      ]}
      searchPlaceholder={t("searchPayslipsPlaceholder")}
      filterLabel={t("tableStatus")}
      allLabel={t("allLabel")}
      ariaLabel={t("payslipsTitle")}
      status={
        payslipsQuery.isLoading
          ? { type: "loading" }
          : payslipsQuery.isError
            ? {
                type: "error",
                message: payslipsQuery.error.message,
                onRetry: () => void payslipsQuery.refetch(),
              }
            : undefined
      }
    />
  );
}
