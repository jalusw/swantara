"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { SubscriptionPlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatInterval } from "../../_components/subscription-utils";
import { SubscriptionPlanFormDialog } from "./subscription-plan-form-dialog";

export function SubscriptionPlansSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Subscriptions");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editPlan, setEditPlan] = useState<SubscriptionPlan | null>(null);

  const plansQuery = useOrgListQuery<{ plans: SubscriptionPlan[] }, Record<string, never>>(
    "subscriptionPlans",
    (organizationId) => getSwantaraService().subscriptionPlans.list(organizationId),
  );

  const plans = plansQuery.data?.plans ?? [];

  const deleteMutation = useMutation({
    mutationFn: (planId: number) =>
      getSwantaraService().subscriptionPlans.delete(Number(orgId), planId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["subscriptionPlans", Number(orgId)] });
    },
    onError: () => {
      toast.error(t("deleteFailed"));
    },
  });

  function handleSave(_id: string) {
    setDialogOpen(false);
    setEditPlan(null);
    void plansQuery.refetch();
  }

  function handleEdit(plan: SubscriptionPlan) {
    setEditPlan(plan);
    setDialogOpen(true);
  }

  const columns: ColumnDef<SubscriptionPlan>[] = [
    {
      accessorKey: "name",
      header: t("colName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "recurringInterval",
      header: t("colInterval"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {formatInterval(row.original.recurringInterval, row.original.recurringCount)}
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
          confirmTitle={t("deletePlanTitle")}
          confirmDescription={t("deletePlanDescription")}
          onEdit={() => handleEdit(row.original)}
          onDelete={() => deleteMutation.mutate(row.original.id)}
        />
      ),
    },
  ];

  const isLoading = plansQuery.isLoading;
  const error = plansQuery.isError ? plansQuery.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={plans}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={t("plansSearchPlaceholder")}
        ariaLabel={t("plansTitle")}
        emptyTitle={t("plansEmpty")}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void plansQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button
            size="sm"
            onClick={() => {
              setEditPlan(null);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{tCommon("add")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <SubscriptionPlanFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initial={editPlan}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
