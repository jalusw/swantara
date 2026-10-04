"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Rma } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { RmaFormDialog } from "./rma-form-dialog";
import { type RmaState, rmaStateLabel, rmaStateTone, rmaTypeLabel } from "./rma-utils";

export function RmasSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Sales");
  const [dialogOpen, setDialogOpen] = useState(false);

  const rmasQuery = useOrgListQuery<{ rmas: Rma[] }, Record<string, never>>(
    "rmas",
    (organizationId) => getSwantaraService().rmas.list(organizationId),
  );

  const rmas = rmasQuery.data?.rmas ?? [];

  function rmaStateLabelText(state: RmaState): string {
    try {
      return (t as unknown as (k: string) => string)(`rmaState.${state}`);
    } catch {
      return rmaStateLabel(state);
    }
  }

  function rmaTypeLabelText(type: Rma["type"]): string {
    try {
      return (t as unknown as (k: string) => string)(`rmaType.${type}`);
    } catch {
      return rmaTypeLabel(type);
    }
  }

  const columns: ColumnDef<Rma>[] = [
    {
      accessorKey: "name",
      header: t("tableNumber"),
      cell: ({ row }) => (
        <a
          href={`/sale-orders/returns/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `RMA-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "type",
      header: t("tableType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{rmaTypeLabelText(row.original.type)}</span>
      ),
    },
    {
      accessorKey: "contactId",
      header: t("fieldContact"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{`#${row.original.contactId}`}</span>
      ),
    },
    {
      accessorKey: "originOrderType",
      header: t("fieldOriginOrderType"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.originOrderType && row.original.originOrderId
            ? `${row.original.originOrderType.toUpperCase()}-${row.original.originOrderId}`
            : "—"}
        </span>
      ),
    },
    {
      accessorKey: "createdAt",
      header: t("tableCreated"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const state = row.original.state as RmaState;
        return (
          <Badge variant="outline" className={rmaStateTone(state)}>
            {rmaStateLabelText(state)}
          </Badge>
        );
      },
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={rmas}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: t("rmaStateDraft") },
          { value: "confirmed", label: t("rmaStateConfirmed") },
          { value: "received", label: t("rmaStateReceived") },
          { value: "refunded", label: t("rmaStateRefunded") },
          { value: "done", label: t("rmaStateDone") },
          { value: "cancelled", label: t("rmaStateCancelled") },
        ]}
        searchPlaceholder={t("searchRmasPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allLabel")}
        ariaLabel={t("rmasTitle")}
        emptyTitle={t("emptyRmas")}
        status={
          rmasQuery.isLoading
            ? { type: "loading" }
            : rmasQuery.isError
              ? {
                  type: "error",
                  message: rmasQuery.error.message,
                  onRetry: () => void rmasQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newRma")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <RmaFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={() => {
            setDialogOpen(false);
            void rmasQuery.refetch();
          }}
        />
      ) : null}
    </div>
  );
}
