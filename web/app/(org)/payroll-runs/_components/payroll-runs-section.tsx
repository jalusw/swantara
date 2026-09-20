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
import type { PayrollRun } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { PayrollRunFormDialog } from "./payroll-run-form-dialog";
import { payrollRunStateTone } from "./payroll-utils";

export function PayrollRunsSection({ orgId }: { orgId: string }) {
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

  const columns: ColumnDef<PayrollRun>[] = [
    {
      accessorKey: "name",
      header: "Name",
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
      header: "Period",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {formatDate(row.original.periodStart)} – {formatDate(row.original.periodEnd)}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
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
          { value: "draft", label: "Draft" },
          { value: "confirmed", label: "Confirmed" },
          { value: "paid", label: "Paid" },
          { value: "closed", label: "Closed" },
        ]}
        searchPlaceholder={"Search runs…"}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"Payroll runs"}
        emptyTitle={"No payroll runs found"}
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
            <span>{"New run"}</span>
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
