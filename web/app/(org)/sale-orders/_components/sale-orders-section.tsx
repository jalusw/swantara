"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, SaleOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { SaleOrderFormDialog } from "./sale-order-form-dialog";

export function SaleOrdersSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Sales");
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialOpportunity = searchParams.get("quotation") ?? searchParams.get("opportunity");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [prefillOpportunity, setPrefillOpportunity] = useState<string | null>(initialOpportunity);

  useEffect(() => {
    setPrefillOpportunity(initialOpportunity);
  }, [initialOpportunity]);

  const ordersQuery = useOrgListQuery<{ orders: SaleOrder[] }, Record<string, never>>(
    "saleOrders",
    (organizationId) => getSwantaraService().saleOrders.list(organizationId),
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

  function handleSave(orderId: string) {
    setDialogOpen(false);
    setPrefillOpportunity(null);
    void ordersQuery.refetch();
    router.push(`/sale-orders/${orderId}`);
  }

  function stateLabel(state: SaleOrder["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`orderState.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<SaleOrder>[] = [
    {
      accessorKey: "name",
      header: t("tableOrder"),
      cell: ({ row }) => (
        <a
          href={`/sale-orders/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `SO-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "contactId",
      header: t("tableCustomer"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {contactMap.get(row.original.contactId) ?? `#${row.original.contactId}`}
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
      accessorKey: "deliveryStatus",
      header: t("tableDelivery"),
      cell: ({ row }) => <Badge variant="outline">{String(row.original.deliveryStatus)}</Badge>,
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
          editLabel={t("editOrder")}
          deleteLabel={t("deleteOrder")}
          confirmTitle={t("deleteOrderTitle")}
          confirmDescription={t("deleteOrderDescription")}
          onEdit={() => router.push(`/sale-orders/${row.original.id}`)}
          onDelete={() => {
            void getSwantaraService()
              .saleOrders.delete(Number(orgId), row.original.id)
              .then(() => void ordersQuery.refetch())
              .catch(() => toast.error(t("saveFailed")));
          }}
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
          { value: "draft", label: t("orderStateDraft") },
          { value: "sent", label: t("orderStateSent") },
          { value: "confirmed", label: t("orderStateConfirmed") },
          { value: "done", label: t("orderStateDone") },
          { value: "cancelled", label: t("orderStateCancelled") },
        ]}
        searchPlaceholder={t("searchOrdersPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allSaleOrders")}
        ariaLabel={t("allSaleOrders")}
        emptyTitle={t("emptyOrders")}
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
          <Button
            size="sm"
            onClick={() => {
              setPrefillOpportunity(initialOpportunity);
              setDialogOpen(true);
            }}
          >
            <Plus />
            <span>{t("newQuotation")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <SaleOrderFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          initialOpportunityId={prefillOpportunity}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
