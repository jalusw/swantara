"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
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
import {
  type MoState,
  moProgress,
  moProgressColor,
  moStateLabel,
  moStateTone,
} from "./production-order-utils";

export function MoSection({ orgId }: { orgId: string }) {
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
      header: "Name",
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
      header: "Item",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {productMap.get(row.original.itemId) ?? `#${row.original.itemId}`}
        </span>
      ),
    },
    {
      accessorKey: "qtyToProduce",
      header: "Qty to Produce",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums ">{row.original.qtyToProduce}</span>,
    },
    {
      accessorKey: "qtyProduced",
      header: "Qty Produced",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{row.original.qtyProduced}</span>,
    },
    {
      accessorKey: "progress",
      header: "Progress",
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
      header: "Origin",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.origin ? String(row.original.origin) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "datePlannedStart",
      header: "Planned Start",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.datePlannedStart ? formatDate(row.original.datePlannedStart) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const state = row.original.state as MoState;
        return (
          <Badge variant="outline" className={moStateTone(state)}>
            {moStateLabel(state)}
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
          { value: "draft", label: "Draft" },
          { value: "confirmed", label: "Confirmed" },
          { value: "planned", label: "Planned" },
          { value: "in_progress", label: "In Progress" },
          { value: "done", label: "Done" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search manufacturing orders..."}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"Manufacturing Orders"}
        emptyTitle={"No manufacturing orders yet"}
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
            <span>{"New Manufacturing Order"}</span>
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
