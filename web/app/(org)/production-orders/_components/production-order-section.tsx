"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Progress } from "@/components/progress";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item, ProductionOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils/style";
import { MoFormDialog } from "./production-order-form-dialog";
import { type MoState, moProgress, moProgressColor, moStateTone } from "./production-order-utils";

export function MoSection({ orgId }: { orgId: string }) {
  const t = useTranslations("ProductionOrders");
  const moState = (state: string) => (t as unknown as (k: string) => string)(`state_${state}`);
  const [dialogOpen, setDialogOpen] = useState(false);

  const moQuery = useOrgListQuery<{ productionOrders: ProductionOrder[] }, Record<string, never>>(
    "productionOrders",
    (organizationId) => getSwantaraService().productionOrders.list(organizationId),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const mos = moQuery.data?.productionOrders ?? [];
  const products = productsQuery.data?.products ?? [];
  const productMap = new Map(products.map((p) => [p.id, p.name]));

  const columns: ColumnDef<ProductionOrder>[] = [
    {
      accessorKey: "name",
      header: t("colName"),
      cell: ({ row }) => (
        <a
          href={`/production-orders/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `MO-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "itemId",
      header: t("colItem"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {productMap.get(row.original.itemId) ?? `#${row.original.itemId}`}
        </span>
      ),
    },
    {
      accessorKey: "qtyToProduce",
      header: t("colQtyToProduce"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.qtyToProduce}</span>,
    },
    {
      accessorKey: "qtyProduced",
      header: t("colQtyProduced"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{row.original.qtyProduced}</span>,
    },
    {
      accessorKey: "progress",
      header: t("colProgress"),
      meta: { align: "right" },
      cell: ({ row }) => {
        const progress = moProgress(row.original);
        return (
          <div className="flex items-center gap-2">
            <Progress value={progress} className="w-16" />
            <span className={cn("text-xs tabular-nums", moProgressColor(progress))}>
              {progress}%
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: "origin",
      header: t("colOrigin"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.origin ? String(row.original.origin) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "datePlannedStart",
      header: t("colPlannedStart"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.datePlannedStart ? formatDate(row.original.datePlannedStart) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => {
        const state = row.original.state as MoState;
        return (
          <Badge variant="outline" className={moStateTone(state)}>
            {moState(state)}
          </Badge>
        );
      },
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={mos}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: moState("draft") },
          { value: "confirmed", label: moState("confirmed") },
          { value: "planned", label: moState("planned") },
          { value: "in_progress", label: moState("in_progress") },
          { value: "done", label: moState("done") },
          { value: "cancelled", label: moState("cancelled") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allOrders")}
        ariaLabel={t("title")}
        emptyTitle={t("emptyTitle")}
        status={
          moQuery.isLoading
            ? { type: "loading" }
            : moQuery.isError
              ? {
                  type: "error",
                  message: moQuery.error.message,
                  onRetry: () => void moQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newOrder")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <MoFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          products={products}
          onSave={() => {
            setDialogOpen(false);
            void moQuery.refetch();
          }}
        />
      ) : null}
    </div>
  );
}
