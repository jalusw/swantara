"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Shipment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { makeStateVariant, nameColumn, stateColumn } from "@/lib/utils/table-columns";

function toRow(shipment: Shipment) {
  return {
    id: String(shipment.id),
    name: shipment.name ?? `PK-${shipment.id}`,
    type: shipment.type,
    state: shipment.state,
    scheduledDate: shipment.scheduledDate,
    origin: shipment.origin,
  };
}

export function ShipmentsSection() {
  const shipmentsQuery = useOrgListQuery<{ shipments: Shipment[] }, Record<string, never>>(
    "stockShipments",
    (organizationId) => getSwantaraService().inventory.stockShipments(organizationId),
  );

  const rows = (shipmentsQuery.data?.shipments ?? []).map(toRow);

  const columns: ColumnDef<ReturnType<typeof toRow>>[] = [
    nameColumn<ReturnType<typeof toRow>>({
      basePath: "stock/shipments",
      header: "Shipment",
    }),
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => <Badge variant="secondary">{String(row.original.type)}</Badge>,
    },
    stateColumn<ReturnType<typeof toRow>>({
      header: "State",
      label: (state) => String(state),
      variant: makeStateVariant(["done"]),
    }),
    {
      accessorKey: "scheduledDate",
      header: "Scheduled",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.scheduledDate
            ? formatDate(row.original.scheduledDate, { nullFallback: "—" })
            : "—"}
        </span>
      ),
    },
    {
      accessorKey: "origin",
      header: "Origin",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.origin ?? "—"}</span>
      ),
    },
  ];

  const isLoading = shipmentsQuery.isLoading;
  const error = shipmentsQuery.isError ? shipmentsQuery.error : null;

  return (
    <InteractiveEntityTable
      columns={columns}
      data={rows}
      getRowId={(row) => row.id}
      searchKeys={["name"]}
      searchPlaceholder={"Search shipments…"}
      ariaLabel={"All shipments"}
      emptyTitle={"No shipments found."}
      status={
        isLoading
          ? { type: "loading" }
          : error
            ? {
                type: "error",
                message: error.message,
                onRetry: () => void shipmentsQuery.refetch(),
              }
            : undefined
      }
    />
  );
}
