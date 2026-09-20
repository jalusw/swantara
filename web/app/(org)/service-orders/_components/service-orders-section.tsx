"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ServiceOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { ServiceOrderFormDialog } from "./service-order-form-dialog";
import { serviceOrderStateTone } from "./service-order-utils";

export function ServiceOrdersSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ serviceOrders: ServiceOrder[] }, Record<string, never>>(
    "serviceOrders",
    (organizationId) => getSwantaraService().serviceOrders.list(organizationId),
  );

  const orders = query.data?.serviceOrders ?? [];

  function handleSave(id: string) {
    setDialogOpen(false);
    void query.refetch();
    router.push(`/service-orders/${id}`);
  }

  const columns: ColumnDef<ServiceOrder>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <a
          href={`/service-orders/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{String(row.original.type)}</span>
      ),
    },
    {
      accessorKey: "priority",
      header: "Priority",
      cell: ({ row }) => <span className="text-sm tabular-nums">{row.original.priority}</span>,
    },
    {
      accessorKey: "scheduledDate",
      header: "Scheduled",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.scheduledDate ? formatDate(String(row.original.scheduledDate)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const tone = serviceOrderStateTone(row.original.state);
        return (
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "warning"
                  ? "border-warning text-warning"
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : tone === "info"
                      ? "border-info text-info"
                      : ""
            }
          >
            {humanizeKey(String(row.original.state))}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"View"}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/service-orders/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader>
            <CardTitle>{"Total Orders"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{orders.length}</p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={orders}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "new", label: "New" },
          { value: "scheduled", label: "Scheduled" },
          { value: "in_progress", label: "In progress" },
          { value: "done", label: "Done" },
          { value: "invoiced", label: "Invoiced" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search service orders…"}
        filterLabel={"State"}
        allLabel={"All service orders"}
        ariaLabel={"Service orders"}
        emptyTitle={"No service orders yet"}
        status={
          query.isLoading
            ? { type: "loading" }
            : query.isError
              ? {
                  type: "error",
                  message: query.error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Create service order"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ServiceOrderFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
