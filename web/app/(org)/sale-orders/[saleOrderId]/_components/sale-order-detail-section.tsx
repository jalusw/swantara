"use client";

import { useState } from "react";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import { JournalDateDialog } from "@/components/journal-date-dialog";
import { PaymentDialog } from "@/components/payment-dialog";
import { RecordLayout } from "@/components/record-layout";
import { StateBadge } from "@/components/state-badge";
import { SaleOrderDeliveryTab } from "./sale-order-delivery-tab";
import { useSaleOrderDetail } from "./sale-order-detail-hooks";
import { SaleOrderInvoicingTab } from "./sale-order-invoicing-tab";
import { SaleOrderOverviewTab } from "./sale-order-overview-tab";
import { SaleOrderPaymentsTab } from "./sale-order-payments-tab";
import { SaleOrderReminderTab } from "./sale-order-reminder-tab";
import { getOrderStatuses } from "./sale-order-status";

type SaleOrderDetailProps = {
  orgId: string;
  saleOrderId: string;
};

export function SaleOrderDetail({ orgId, saleOrderId }: SaleOrderDetailProps) {
  const [activeDialog, setActiveDialog] = useState<"deliver" | "invoice" | "pay" | null>(null);
  const detail = useSaleOrderDetail(orgId, saleOrderId);
  const { order, isLoading, journals } = detail;

  if (isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!order) {
    return <p className="text-sm text-muted-foreground">{"Sale order not found."}</p>;
  }

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: "All sale orders", href: "/sale-orders" },
          { label: order.name ?? `SO-${order.id}` },
        ]}
        title={order.name ?? `SO-${order.id}`}
        status={<StateBadge value={order.state} statuses={getOrderStatuses()} />}
        tabs={[
          {
            id: "overview",
            label: "Overview",
            content: (
              <SaleOrderOverviewTab
                order={order}
                contactMap={detail.contactMap}
                priceBookMap={detail.priceBookMap}
                warehouseMap={detail.warehouseMap}
                onSend={detail.send}
                onConfirm={detail.confirm}
                onCancel={detail.cancel}
                onDone={detail.markDone}
              />
            ),
          },
          {
            id: "delivery",
            label: "Delivery",
            content: (
              <SaleOrderDeliveryTab
                order={order}
                linkedShipment={detail.linkedShipment}
                onDeliver={() => setActiveDialog("deliver")}
              />
            ),
          },
          {
            id: "invoicing",
            label: "Invoicing",
            content: (
              <SaleOrderInvoicingTab
                deliveredUnbilledLines={detail.deliveredUnbilledLines}
                onInvoice={() => setActiveDialog("invoice")}
              />
            ),
          },
          {
            id: "payments",
            label: "Payments",
            content: <SaleOrderPaymentsTab order={order} onPay={() => setActiveDialog("pay")} />,
          },
          {
            id: "reminder",
            label: "Reminder",
            content: (
              <SaleOrderReminderTab
                reminders={detail.reminders}
                onGenerate={detail.generateReminder}
              />
            ),
          },
        ]}
      />
      <JournalDateDialog
        title={"Ship / Deliver"}
        description={"Pick, pack and ship the order. Validates moves and posts COGS."}
        confirmLabel={"Ship / Deliver"}
        cancelLabel={"Cancel"}
        open={activeDialog === "deliver"}
        onOpenChange={(open) => setActiveDialog(open ? "deliver" : null)}
        onConfirm={(journalId, date) => {
          detail.deliver({ journalId, date });
          setActiveDialog(null);
        }}
        journals={journals}
      />
      <JournalDateDialog
        title={"Create invoice"}
        description={"Invoice delivered-unbilled lines. Shows triplet and residual."}
        confirmLabel={"Create invoice"}
        cancelLabel={"Cancel"}
        open={activeDialog === "invoice"}
        onOpenChange={(open) => setActiveDialog(open ? "invoice" : null)}
        onConfirm={(journalId, date) => {
          detail.invoice({ journalId, date });
          setActiveDialog(null);
        }}
        journals={journals}
      />
      <PaymentDialog
        title={"Collect payment"}
        description={"Capture payment and allocate to invoices. Shows payment_state and residual."}
        confirmLabel={"Collect payment"}
        cancelLabel={"Cancel"}
        open={activeDialog === "pay"}
        onOpenChange={(open) => setActiveDialog(open ? "pay" : null)}
        onConfirm={(values) => {
          detail.pay(values as { journalId: string; amount: string; date: string });
          setActiveDialog(null);
        }}
        journals={journals}
        showAmount
      />
    </>
  );
}
