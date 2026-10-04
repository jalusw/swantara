"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ExpenseReport } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { ExpenseFormDialog } from "./expense-form-dialog";
import { expenseStateTone } from "./expense-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function ExpensesSection({ orgId }: { orgId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Expenses");
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
      header: () => t("fieldName"),
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
      header: () => t("fieldEmployee"),
      cell: ({ row }) => <span className="text-muted-foreground">#{row.original.employeeId}</span>,
    },
    {
      accessorKey: "paymentMode",
      header: () => t("fieldPaymentMode"),
      cell: ({ row }) => (
        <Badge variant="outline">
          {(t as unknown as (k: string) => string)(`paymentMode_${row.original.paymentMode}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "totalAmount",
      header: () => t("colTotal"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.totalAmount)}</span>
      ),
    },
    {
      accessorKey: "state",
      header: () => t("colStatus"),
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
            {(t as unknown as (k: string) => string)(`expenseState_${row.original.state}`)}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("view")}
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
          { value: "draft", label: (t as unknown as (k: string) => string)("expenseState_draft") },
          {
            value: "submitted",
            label: (t as unknown as (k: string) => string)("expenseState_submitted"),
          },
          {
            value: "approved",
            label: (t as unknown as (k: string) => string)("expenseState_approved"),
          },
          {
            value: "refused",
            label: (t as unknown as (k: string) => string)("expenseState_refused"),
          },
          {
            value: "posted",
            label: (t as unknown as (k: string) => string)("expenseState_posted"),
          },
          {
            value: "reimbursed",
            label: (t as unknown as (k: string) => string)("expenseState_reimbursed"),
          },
        ]}
        searchPlaceholder={t("searchExpenses")}
        filterLabel={t("colStatus")}
        allLabel={t("allExpenses")}
        ariaLabel={t("expensesTitle")}
        emptyTitle={t("expensesEmpty")}
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
            <span>{t("newReport")}</span>
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
