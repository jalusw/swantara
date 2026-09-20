"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ReorderRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { activeColumn } from "@/lib/utils/table-columns";
import { ReorderRuleFormDialog } from "./reorder-rule-form-dialog";

function toRow(rule: ReorderRule) {
  return {
    id: String(rule.id),
    itemId: rule.itemId,
    minQty: rule.minQty,
    maxQty: rule.maxQty,
    qtyMultiple: rule.qtyMultiple,
    leadTimeDays: rule.leadTimeDays,
    active: rule.active,
  };
}

export function ReorderRulesSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<ReorderRule | null>(null);

  const rulesQuery = useOrgListQuery<{ rules: ReorderRule[] }, Record<string, never>>(
    "reorderRules",
    (organizationId) => getSwantaraService().inventory.reorderRules(organizationId),
  );

  const rows = (rulesQuery.data?.rules ?? []).map(toRow);

  function handleSave() {
    setDialogOpen(false);
    setEditing(null);
    void rulesQuery.refetch();
  }

  function handleDelete(rule: ReturnType<typeof toRow>) {
    void getSwantaraService()
      .inventory.deleteReorderRule(Number(orgId), Number(rule.id))
      .then(() => void rulesQuery.refetch())
      .catch(() => toast.error("Could not disable the organization."));
  }

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    {
      accessorKey: "itemId",
      header: "Item",
      cell: ({ row }) => <span>{row.original.itemId}</span>,
    },
    {
      accessorKey: "minQty",
      header: "Min qty",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.minQty)}</span>,
    },
    {
      accessorKey: "maxQty",
      header: "Max qty",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.maxQty)}</span>,
    },
    {
      accessorKey: "qtyMultiple",
      header: "Qty multiple",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">{formatNumber(row.original.qtyMultiple)}</span>
      ),
    },
    {
      accessorKey: "leadTimeDays",
      header: "Lead time (days)",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.leadTimeDays != null ? `${row.original.leadTimeDays}d` : "—"}
        </span>
      ),
    },
    activeColumn<ReturnType<typeof toRow>>({
      header: "Active",
      activeLabel: "Active",
      inactiveLabel: "Inactive",
    }),
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete this reorder rule?"}
          confirmDescription={"This will remove the automatic replenishment rule for this item."}
          onEdit={() => {
            const rule = (rulesQuery.data?.rules ?? []).find(
              (r) => String(r.id) === row.original.id,
            );
            if (rule) {
              setEditing(rule);
              setDialogOpen(true);
            }
          }}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  const isLoading = rulesQuery.isLoading;
  const error = rulesQuery.isError ? rulesQuery.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={rows}
        getRowId={(row) => row.id}
        searchKeys={["itemId"]}
        searchPlaceholder={"Search reorder rules…"}
        ariaLabel={"All reorder rules"}
        emptyTitle={"No reorder rules found. Create one to set up automatic replenishment."}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void rulesQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button
            size="sm"
            onClick={() => {
              setEditing(null);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{"Add rule"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ReorderRuleFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initial={editing}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
