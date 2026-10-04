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
import type { PayrollRun } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { PayrollRunFormDialog } from "./payroll-run-form-dialog";
import { payrollRunStateTone } from "./payroll-utils";

export function PayrollRunsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Payroll");
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const runsQuery = useOrgListQuery<{ payrollRuns: PayrollRun[] }, Record<string, never>>(
    "payrollRuns",
    (organizationId) => getSwantaraService().payrollRuns.list(organizationId),
  );

  const runs = runsQuery.data?.payrollRuns ?? [];

  function handleSave(runId: string) {
    setDialogOpen(false);
    void runsQuery.refetch();
    router.push(`/payroll-runs/${runId}`);
  }

  function stateLabel(state: PayrollRun["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`runState.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<PayrollRun>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => (
        <a
          href={`/payroll-runs/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `PR-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "periodStart",
      header: t("tablePeriod"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {formatDate(row.original.periodStart)} – {formatDate(row.original.periodEnd)}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const tone = payrollRunStateTone(row.original.state);
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
            {stateLabel(row.original.state)}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("actionView")}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/payroll-runs/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={runs}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: t("runStateDraft") },
          { value: "confirmed", label: t("runStateConfirmed") },
          { value: "paid", label: t("runStatePaid") },
          { value: "closed", label: t("runStateClosed") },
        ]}
        searchPlaceholder={t("searchRunsPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allLabel")}
        ariaLabel={t("title")}
        emptyTitle={t("emptyRuns")}
        status={
          runsQuery.isLoading
            ? { type: "loading" }
            : runsQuery.isError
              ? {
                  type: "error",
                  message: runsQuery.error.message,
                  onRetry: () => void runsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newRun")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <PayrollRunFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
