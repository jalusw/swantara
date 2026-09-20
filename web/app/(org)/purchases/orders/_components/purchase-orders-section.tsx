"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, PurchaseOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { PurchaseOrderFormDialog } from "./purchase-order-form-dialog";

export function PurchaseOrdersSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const ordersQuery = useOrgListQuery<{ orders: PurchaseOrder[] }, Record<string, never>>(
    "purchaseOrders",
    (organizationId) => getSwantaraService().purchaseOrders.list(organizationId),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const orders = ordersQuery.data?.orders ?? [];
  const contacts = contactsQuery.data?.contacts ?? [];
  const contactMap = useMemo(
    () => new Map(contacts.map((p) => [p.id, p.displayName || p.name])),
    [contacts],
  );

  function handleSave() {
    setDialogOpen(false);
    void ordersQuery.refetch();
  }

  const columns: ColumnDef<PurchaseOrder>[] = [
    {
      accessorKey: "name",
      header: "Order",
      cell: ({ row }) => (
        <a
          href={`/purchases/orders/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `PO-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "supplierId",
      header: "Supplier",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {contactMap.get(row.original.supplierId) ?? `#${row.original.supplierId}`}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ row }) => (
        <Badge
          variant={
            row.original.state === "done"
              ? "default"
              : row.original.state === "cancelled"
                ? "outline"
                : "secondary"
          }
        >
          {humanizeKey(String(row.original.state))}
        </Badge>
      ),
    },
    {
      accessorKey: "receiptStatus",
      header: "Receipt",
      cell: ({ row }) => <Badge variant="outline">{String(row.original.receiptStatus)}</Badge>,
    },
    {
      accessorKey: "invoiceStatus",
      header: "Invoicing",
      cell: ({ row }) => <Badge variant="outline">{String(row.original.invoiceStatus)}</Badge>,
    },
    {
      accessorKey: "amountTotal",
      header: "Total",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.amountTotal)}</span>
      ),
    },
    {
      accessorKey: "orderDate",
      header: "Order date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.orderDate ? formatDate(row.original.orderDate) : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit purchase order"}
          deleteLabel={"Delete purchase order"}
          confirmTitle={"Delete this purchase order?"}
          confirmDescription={"The purchase order will be removed. This cannot be undone."}
          onEdit={() => router.push(`/purchases/orders/${row.original.id}`)}
        />
      ),
    },
  ];

  const isLoading = ordersQuery.isLoading || contactsQuery.isLoading;
  const error = ordersQuery.isError
    ? ordersQuery.error
    : contactsQuery.isError
      ? contactsQuery.error
      : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={orders}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "sent", label: "Sent" },
          { value: "confirmed", label: "Confirmed" },
          { value: "done", label: "Done" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search purchase orders…"}
        filterLabel={"Status"}
        allLabel={"All purchase orders"}
        ariaLabel={"All purchase orders"}
        emptyTitle={"No purchase orders"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void ordersQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New purchase order"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <PurchaseOrderFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
