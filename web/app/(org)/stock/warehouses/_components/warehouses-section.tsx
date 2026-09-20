"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Warehouse } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { WarehouseFormDialog } from "./warehouse-form-dialog";

function toRow(warehouse: Warehouse) {
  return {
    id: String(warehouse.id),
    name: warehouse.name,
    code: warehouse.code,
    createdAt: warehouse.createdAt,
  };
}

export function WarehousesSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const warehousesQuery = useOrgListQuery<{ warehouses: Warehouse[] }, Record<string, never>>(
    "warehouses",
    (organizationId) => getSwantaraService().inventory.warehouses(organizationId),
  );

  const rows = (warehousesQuery.data?.warehouses ?? []).map(toRow);

  function handleSave(_warehouseId: string) {
    setDialogOpen(false);
    void warehousesQuery.refetch();
  }

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <Link
          href={`/stock/warehouses/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </Link>
      ),
    },
    {
      accessorKey: "code",
      header: "Code",
      cell: ({ row }) => (
        <span className="font-mono text-sm text-muted-foreground">{row.original.code ?? "—"}</span>
      ),
    },
  ];

  const isLoading = warehousesQuery.isLoading;
  const error = warehousesQuery.isError ? warehousesQuery.error : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={rows}
        getRowId={(row) => row.id}
        searchKeys={["name"]}
        searchPlaceholder={"Search warehouses…"}
        ariaLabel={"All warehouses"}
        emptyTitle={"No warehouses found. Create one to get started."}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => {
                    void warehousesQuery.refetch();
                  },
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Add warehouse"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <WarehouseFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
