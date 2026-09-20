"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Payslip } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

export function PayslipsSection() {
  const payslipsQuery = useOrgListQuery<{ payslips: Payslip[] }, Record<string, never>>(
    "payslips",
    (organizationId) => getSwantaraService().payslips.list(organizationId),
  );

  const payslips = payslipsQuery.data?.payslips ?? [];

  const columns: ColumnDef<Payslip>[] = [
    {
      accessorKey: "id",
      header: "Number",
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
      header: "Employee",
      cell: ({ row }) => <span className="text-muted-foreground">#{row.original.employeeId}</span>,
    },
    {
      accessorKey: "gross",
      header: "Gross",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatNumber(row.original.gross)}</span>,
    },
    {
      accessorKey: "net",
      header: "Net",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{formatNumber(row.original.net)}</span>,
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <Badge variant={row.original.state === "posted" ? "default" : "outline"}>
          {humanizeKey(String(row.original.state))}
        </Badge>
      ),
    },
    {
      accessorKey: "moveId",
      header: "Journal entry",
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
        { value: "draft", label: "Draft" },
        { value: "posted", label: "Posted" },
      ]}
      searchPlaceholder={"Search payslips…"}
      filterLabel={"State"}
      allLabel={"All"}
      ariaLabel={"Payslips"}
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
