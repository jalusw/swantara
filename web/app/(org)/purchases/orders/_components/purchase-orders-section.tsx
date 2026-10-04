"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Purchases");
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

  function stateLabel(state: PurchaseOrder["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`poState.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<PurchaseOrder>[] = [
    {
      accessorKey: "name",
      header: t("tableOrder"),
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
      header: t("tableSupplier"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {contactMap.get(row.original.supplierId) ?? `#${row.original.supplierId}`}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("tableStatus"),
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
          {stateLabel(row.original.state)}
        </Badge>
      ),
    },
    {
      accessorKey: "receiptStatus",
      header: t("tableReceipt"),
      cell: ({ row }) => <Badge variant="outline">{String(row.original.receiptStatus)}</Badge>,
    },
    {
      accessorKey: "invoiceStatus",
      header: t("tableInvoicing"),
      cell: ({ row }) => <Badge variant="outline">{String(row.original.invoiceStatus)}</Badge>,
    },
    {
      accessorKey: "amountTotal",
      header: t("tableTotal"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.amountTotal)}</span>
      ),
    },
    {
      accessorKey: "orderDate",
      header: t("tableOrderDate"),
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
          editLabel={t("editPurchaseOrder")}
          deleteLabel={t("deletePurchaseOrder")}
          confirmTitle={t("deletePurchaseOrderTitle")}
          confirmDescription={t("deletePurchaseOrderDescription")}
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
          { value: "draft", label: t("poStateDraft") },
          { value: "sent", label: t("poStateSent") },
          { value: "confirmed", label: t("poStateConfirmed") },
          { value: "done", label: t("poStateDone") },
          { value: "cancelled", label: t("poStateCancelled") },
        ]}
        searchPlaceholder={t("searchPurchaseOrdersPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allPurchaseOrders")}
        ariaLabel={t("allPurchaseOrders")}
        emptyTitle={t("emptyPurchaseOrders")}
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
            <span>{t("newPurchaseOrder")}</span>
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
