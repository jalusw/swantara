"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CommissionAssignment, CommissionPlan, CommissionRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { CommissionAssignmentFormDialog } from "./commission-assignment-form-dialog";
import { CommissionRuleFormDialog } from "./commission-rule-form-dialog";

export function CommissionPlanDetail({ orgId, planId }: { orgId: string; planId: string }) {
  const [ruleDialogOpen, setRuleDialogOpen] = useState(false);
  const [assignmentDialogOpen, setAssignmentDialogOpen] = useState(false);

  const planQuery = useOrgListQuery<{ commissionPlan: CommissionPlan }, Record<string, never>>(
    "commissionPlan",
    (organizationId) => getSwantaraService().commissionPlans.get(organizationId, Number(planId)),
  );

  const rulesQuery = useOrgListQuery<{ commissionRules: CommissionRule[] }, Record<string, never>>(
    "commissionRules",
    (organizationId) =>
      getSwantaraService().commissionPlans.rules.list(organizationId, Number(planId)),
  );

  const assignmentsQuery = useOrgListQuery<
    { commissionAssignments: CommissionAssignment[] },
    Record<string, never>
  >("commissionAssignments", (organizationId) =>
    getSwantaraService().commissionPlans.assignments.list(organizationId, Number(planId)),
  );

  const plan = planQuery.data?.commissionPlan;
  const rules = rulesQuery.data?.commissionRules ?? [];
  const assignments = assignmentsQuery.data?.commissionAssignments ?? [];

  function handleRuleSaved() {
    setRuleDialogOpen(false);
    void rulesQuery.refetch();
  }

  function handleAssignmentSaved() {
    setAssignmentDialogOpen(false);
    void assignmentsQuery.refetch();
  }

  if (planQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!plan) {
    return <p className="text-sm text-muted-foreground">{"Commission plan not found."}</p>;
  }

  const ruleColumns: ColumnDef<CommissionRule>[] = [
    {
      accessorKey: "itemCategoryId",
      header: "Category",
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.itemCategoryId ? `#${row.original.itemCategoryId}` : "All Categories"}
        </span>
      ),
    },
    {
      accessorKey: "minAmount",
      header: "Min",
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.minAmount}</span>,
    },
    {
      accessorKey: "maxAmount",
      header: "Max",
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.maxAmount}</span>,
    },
    {
      accessorKey: "ratePct",
      header: "Rate %",
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.ratePct}%</span>,
    },
    {
      accessorKey: "fixedAmount",
      header: "Fixed",
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.fixedAmount}</span>,
    },
  ];

  const assignmentColumns: ColumnDef<CommissionAssignment>[] = [
    {
      accessorKey: "salespersonId",
      header: "Salesperson ID",
      cell: ({ row }) => <span className="text-sm">#{row.original.salespersonId}</span>,
    },
    {
      accessorKey: "dateStart",
      header: "Start date",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(row.original.dateStart, { nullFallback: "—" })}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: "End date",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(row.original.dateEnd, { nullFallback: "—" })}
        </span>
      ),
    },
  ];

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Commission plans", href: "/commission-plans" },
        { label: plan.name },
      ]}
      title={plan.name}
      status={
        <Badge
          variant="outline"
          className={
            plan.active
              ? "border-success text-success"
              : "border-muted-foreground text-muted-foreground"
          }
        >
          {plan.active ? "Active" : "Inactive"}
        </Badge>
      }
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Basis"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{humanizeKey(String(plan.basis))}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Total rules"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">{rules.length}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Total assignments"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">{assignments.length}</p>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "rules",
          label: "Rules",
          content: (
            <>
              <InteractiveEntityTable
                columns={ruleColumns}
                data={rules}
                getRowId={(row) => String(row.id)}
                searchKeys={["itemCategoryId"]}
                searchPlaceholder={"Search rules…"}
                ariaLabel={"Rules"}
                emptyTitle={"No rules configured."}
                actions={
                  <Button size="sm" onClick={() => setRuleDialogOpen(true)}>
                    <Plus />
                    <span>{"Add rule"}</span>
                  </Button>
                }
              />
              {ruleDialogOpen ? (
                <CommissionRuleFormDialog
                  open={ruleDialogOpen}
                  onOpenChange={setRuleDialogOpen}
                  orgId={orgId}
                  planId={planId}
                  onSave={handleRuleSaved}
                />
              ) : null}
            </>
          ),
        },
        {
          id: "assignments",
          label: "Assignments",
          content: (
            <>
              <InteractiveEntityTable
                columns={assignmentColumns}
                data={assignments}
                getRowId={(row) => String(row.id)}
                searchKeys={["salespersonId"]}
                searchPlaceholder={"Search assignments…"}
                ariaLabel={"Assignments"}
                emptyTitle={"No assignments."}
                actions={
                  <Button size="sm" onClick={() => setAssignmentDialogOpen(true)}>
                    <Plus />
                    <span>{"Add assignment"}</span>
                  </Button>
                }
              />
              {assignmentDialogOpen ? (
                <CommissionAssignmentFormDialog
                  open={assignmentDialogOpen}
                  onOpenChange={setAssignmentDialogOpen}
                  orgId={orgId}
                  planId={planId}
                  onSave={handleAssignmentSaved}
                />
              ) : null}
            </>
          ),
        },
      ]}
    />
  );
}
