"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Account, SalaryRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { SalaryRuleFormDialog } from "./salary-rule-form-dialog";

export function SalaryRulesSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Payroll");
  const tCommon = useTranslations("Common");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<SalaryRule | null>(null);

  const rulesQuery = useOrgListQuery<{ salaryRules: SalaryRule[] }, Record<string, never>>(
    "salaryRules",
    (organizationId) => getSwantaraService().salaryRules.list(organizationId),
  );

  const accountsQuery = useOrgListQuery<{ accounts: Account[] }, Record<string, never>>(
    "accounts",
    (organizationId) => getSwantaraService().accounts.list(organizationId),
  );

  const rules = rulesQuery.data?.salaryRules ?? [];
  const accounts = accountsQuery.data?.accounts ?? [];

  function handleSave() {
    setDialogOpen(false);
    setEditingRule(null);
    void rulesQuery.refetch();
  }

  function handleEdit(rule: SalaryRule) {
    setEditingRule(rule);
    setDialogOpen(true);
  }

  function handleDelete(rule: SalaryRule) {
    void getSwantaraService()
      .salaryRules.delete(Number(orgId), rule.id)
      .then(() => void rulesQuery.refetch());
  }

  const columns: ColumnDef<SalaryRule>[] = [
    {
      accessorKey: "code",
      header: t("tableCode"),
      cell: ({ row }) => <span className="">{row.original.code}</span>,
    },
    {
      accessorKey: "name",
      header: t("tableName"),
    },
    {
      accessorKey: "category",
      header: t("tableCategory"),
      cell: ({ row }) => (
        <Badge variant={row.original.category === "earning" ? "default" : "destructive"}>
          {String(row.original.category ?? "earning")}
        </Badge>
      ),
    },
    {
      accessorKey: "computeType",
      header: t("tableComputeType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{String(row.original.computeType ?? "fixed")}</span>
      ),
    },
    {
      accessorKey: "amount",
      header: t("tableAmount"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.amount != null ? row.original.amount : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={tCommon("edit")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteSalaryRuleTitle")}
          confirmDescription={t("deleteSalaryRuleDescription")}
          onEdit={() => handleEdit(row.original)}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={rules}
        getRowId={(row) => String(row.id)}
        searchKeys={["code", "name"]}
        statusKey="category"
        statusOptions={[
          { value: "earning", label: t("categoryEarning") },
          { value: "deduction", label: t("categoryDeduction") },
        ]}
        searchPlaceholder={t("searchSalaryRulesPlaceholder")}
        filterLabel={t("tableCategory")}
        allLabel={t("allLabel")}
        ariaLabel={t("salaryRulesTitle")}
        status={
          rulesQuery.isLoading
            ? { type: "loading" }
            : rulesQuery.isError
              ? {
                  type: "error",
                  message: rulesQuery.error.message,
                  onRetry: () => void rulesQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button
            size="sm"
            onClick={() => {
              setEditingRule(null);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{t("addSalaryRule")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <SalaryRuleFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initial={editingRule}
          accounts={accounts}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
