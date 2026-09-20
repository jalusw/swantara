"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
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
  const [dialogOpen, setDialogOpen] = useState(false);

  const rmasQuery = useOrgListQuery<{ rmas: Rma[] }, Record<string, never>>(
    "rmas",
    (organizationId) => getSwantaraService().rmas.list(organizationId),
  );

  const rmas = rmasQuery.data?.rmas ?? [];

  const columns: ColumnDef<Rma>[] = [
    {
      accessorKey: "name",
      header: "Number",
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
      header: "Type",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{rmaTypeLabel(row.original.type)}</span>
      ),
    },
    {
      accessorKey: "contactId",
      header: "Contact",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{`#${row.original.contactId}`}</span>
      ),
    },
    {
      accessorKey: "originOrderType",
      header: "Origin order",
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
      header: "Created",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const state = row.original.state as RmaState;
        return (
          <Badge variant="outline" className={rmaStateTone(state)}>
            {rmaStateLabel(state)}
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
          { value: "draft", label: "Draft" },
          { value: "confirmed", label: "Confirmed" },
          { value: "received", label: "Received" },
          { value: "refunded", label: "Refunded" },
          { value: "done", label: "Done" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search RMAs..."}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"Returns & RMA"}
        emptyTitle={"No RMAs found."}
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
            <span>{"New RMA"}</span>
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
