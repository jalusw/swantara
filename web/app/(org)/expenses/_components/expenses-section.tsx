"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ExpenseReport } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { ExpenseFormDialog } from "./expense-form-dialog";
import { expenseStateTone } from "./expense-utils";

export function ExpensesSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const reportsQuery = useOrgListQuery<{ reports: ExpenseReport[] }, Record<string, never>>(
    "expenseReports",
    (organizationId) => getSwantaraService().expenseReports.list(organizationId),
  );

  const reports = reportsQuery.data?.reports ?? [];

  function handleSave() {
    setDialogOpen(false);
    void reportsQuery.refetch();
  }

  const columns: ColumnDef<ExpenseReport>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <a
          href={`/expenses/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "employeeId",
      header: "Employee",
      cell: ({ row }) => <span className="text-muted-foreground">#{row.original.employeeId}</span>,
    },
    {
      accessorKey: "paymentMode",
      header: "Payment mode",
      cell: ({ row }) => <Badge variant="outline">{String(row.original.paymentMode)}</Badge>,
    },
    {
      accessorKey: "totalAmount",
      header: "Total",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.totalAmount)}</span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const tone = expenseStateTone(row.original.state);
        return (
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "warning"
                  ? "border-warning text-warning"
                  : tone === "info"
                    ? "border-info text-info"
                    : tone === "danger"
                      ? "border-destructive text-destructive"
                      : ""
            }
          >
            {humanizeKey(String(row.original.state))}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"View"}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/expenses/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={reports}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "submitted", label: "Submitted" },
          { value: "approved", label: "Approved" },
          { value: "refused", label: "Refused" },
          { value: "posted", label: "Posted" },
          { value: "reimbursed", label: "Reimbursed" },
        ]}
        searchPlaceholder={"Search expenses…"}
        filterLabel={"State"}
        allLabel={"All expenses"}
        ariaLabel={"Expenses"}
        emptyTitle={"No expense reports yet"}
        status={
          reportsQuery.isLoading
            ? { type: "loading" }
            : reportsQuery.isError
              ? {
                  type: "error",
                  message: reportsQuery.error.message,
                  onRetry: () => void reportsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Create expense report"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ExpenseFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
