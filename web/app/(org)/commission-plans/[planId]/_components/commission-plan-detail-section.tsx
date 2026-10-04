"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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
import { CommissionAssignmentFormDialog } from "./commission-assignment-form-dialog";
import { CommissionRuleFormDialog } from "./commission-rule-form-dialog";

export function CommissionPlanDetail({ orgId, planId }: { orgId: string; planId: string }) {
  const t = useTranslations("Commissions");
  const tCommon = useTranslations("Common");
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

  function basisLabel(value: string): string {
    return (t as unknown as (k: string) => string)(`basis_${value}`);
  }

  function handleRuleSaved() {
    setRuleDialogOpen(false);
    void rulesQuery.refetch();
  }

  function handleAssignmentSaved() {
    setAssignmentDialogOpen(false);
    void assignmentsQuery.refetch();
  }

  if (planQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!plan) {
    return <p className="text-sm text-muted-foreground">{t("planNotFound")}</p>;
  }

  const ruleColumns: ColumnDef<CommissionRule>[] = [
    {
      accessorKey: "itemCategoryId",
      header: t("tableCategory"),
      cell: ({ row }) => (
        <span className="text-sm">
          {row.original.itemCategoryId ? `#${row.original.itemCategoryId}` : t("allCategories")}
        </span>
      ),
    },
    {
      accessorKey: "minAmount",
      header: t("ruleMin"),
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.minAmount}</span>,
    },
    {
      accessorKey: "maxAmount",
      header: t("ruleMax"),
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.maxAmount}</span>,
    },
    {
      accessorKey: "ratePct",
      header: t("ruleRate"),
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.ratePct}%</span>,
    },
    {
      accessorKey: "fixedAmount",
      header: t("ruleFixed"),
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.fixedAmount}</span>,
    },
  ];

  const assignmentColumns: ColumnDef<CommissionAssignment>[] = [
    {
      accessorKey: "salespersonId",
      header: t("fieldSalespersonId"),
      cell: ({ row }) => <span className="text-sm">#{row.original.salespersonId}</span>,
    },
    {
      accessorKey: "dateStart",
      header: t("fieldDateStart"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(row.original.dateStart, { nullFallback: "—" })}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: t("fieldDateEnd"),
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
        { label: t("plansLabel"), href: "/commission-plans" },
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
          {plan.active ? t("active") : t("inactive")}
        </Badge>
      }
      tabs={[
        {
          id: "overview",
          label: t("tabOverview"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("basis")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{basisLabel(String(plan.basis))}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("totalRules")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">{rules.length}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("totalAssignments")}</CardTitle>
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
          label: t("tabRules"),
          content: (
            <>
              <InteractiveEntityTable
                columns={ruleColumns}
                data={rules}
                getRowId={(row) => String(row.id)}
                searchKeys={["itemCategoryId"]}
                searchPlaceholder={t("searchRulesPlaceholder")}
                ariaLabel={t("rulesLabel")}
                emptyTitle={t("emptyRules")}
                actions={
                  <Button size="sm" onClick={() => setRuleDialogOpen(true)}>
                    <Plus />
                    <span>{t("addRule")}</span>
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
          label: t("tabAssignments"),
          content: (
            <>
              <InteractiveEntityTable
                columns={assignmentColumns}
                data={assignments}
                getRowId={(row) => String(row.id)}
                searchKeys={["salespersonId"]}
                searchPlaceholder={t("searchAssignmentsPlaceholder")}
                ariaLabel={t("assignmentsLabel")}
                emptyTitle={t("emptyAssignments")}
                actions={
                  <Button size="sm" onClick={() => setAssignmentDialogOpen(true)}>
                    <Plus />
                    <span>{t("addAssignment")}</span>
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
