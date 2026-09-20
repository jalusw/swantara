"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
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
      header: "Code",
      cell: ({ row }) => <span className="">{row.original.code}</span>,
    },
    {
      accessorKey: "name",
      header: "Name",
    },
    {
      accessorKey: "category",
      header: "Category",
      cell: ({ row }) => (
        <Badge variant={row.original.category === "earning" ? "default" : "destructive"}>
          {String(row.original.category ?? "earning")}
        </Badge>
      ),
    },
    {
      accessorKey: "computeType",
      header: "Compute type",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{String(row.original.computeType ?? "fixed")}</span>
      ),
    },
    {
      accessorKey: "amount",
      header: "Amount",
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
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete rule"}
          confirmDescription={"Are you sure you want to delete this salary rule?"}
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
          { value: "earning", label: "Earning" },
          { value: "deduction", label: "Deduction" },
        ]}
        searchPlaceholder={"Search rules…"}
        filterLabel={"Category"}
        allLabel={"All"}
        ariaLabel={"Salary rules"}
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
            <span>{"Add rule"}</span>
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
