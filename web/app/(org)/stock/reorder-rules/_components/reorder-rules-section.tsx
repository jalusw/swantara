"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Stock");
  const tCommon = useTranslations("Common");
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
      .catch(() => toast.error(t("toastFailed")));
  }

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    {
      accessorKey: "itemId",
      header: () => t("fieldItem"),
      cell: ({ row }) => <span>{row.original.itemId}</span>,
    },
    {
      accessorKey: "minQty",
      header: () => t("fieldMinQty"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.minQty)}</span>,
    },
    {
      accessorKey: "maxQty",
      header: () => t("fieldMaxQty"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.maxQty)}</span>,
    },
    {
      accessorKey: "qtyMultiple",
      header: () => t("fieldQtyMultiple"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">{formatNumber(row.original.qtyMultiple)}</span>
      ),
    },
    {
      accessorKey: "leadTimeDays",
      header: () => t("fieldLeadTime"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.leadTimeDays != null ? `${row.original.leadTimeDays}d` : "—"}
        </span>
      ),
    },
    activeColumn<ReturnType<typeof toRow>>({
      header: t("active"),
      activeLabel: t("active"),
      inactiveLabel: t("inactive"),
    }),
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={tCommon("edit")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteReorderRuleTitle")}
          confirmDescription={t("deleteReorderRuleDescription")}
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
        searchPlaceholder={t("searchReorderRules")}
        ariaLabel={t("allReorderRules")}
        emptyTitle={t("reorderRulesEmpty")}
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
            <span>{t("addRule")}</span>
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
