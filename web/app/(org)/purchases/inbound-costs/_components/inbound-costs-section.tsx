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
import type { InboundCost } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { InboundCostFormDialog } from "./inbound-cost-form-dialog";
import { inboundCostStateTone } from "./inbound-cost-utils";

export function InboundCostsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Purchases");
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const costsQuery = useOrgListQuery<{ costs: InboundCost[] }, Record<string, never>>(
    "inboundCosts",
    (organizationId) => getSwantaraService().inventory.inboundCosts(organizationId),
  );

  const costs = costsQuery.data?.costs ?? [];

  function handleSave() {
    setDialogOpen(false);
    void costsQuery.refetch();
  }

  function stateLabel(state: InboundCost["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`inboundCostState.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<InboundCost>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => (
        <a
          href={`/purchases/inbound-costs/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "date",
      header: t("tableDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.date ? formatDate(String(row.original.date)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const tone = inboundCostStateTone(row.original.state);
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
          onEdit={() => router.push(`/purchases/inbound-costs/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={costs}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: t("inboundCostStateDraft") },
          { value: "posted", label: t("inboundCostStatePosted") },
          { value: "cancelled", label: t("inboundCostStateCancelled") },
        ]}
        searchPlaceholder={t("searchInboundCostsPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allInboundCosts")}
        ariaLabel={t("inboundCostsTitle")}
        emptyTitle={t("emptyInboundCosts")}
        status={
          costsQuery.isLoading
            ? { type: "loading" }
            : costsQuery.isError
              ? {
                  type: "error",
                  message: costsQuery.error.message,
                  onRetry: () => void costsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newInboundCost")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <InboundCostFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
