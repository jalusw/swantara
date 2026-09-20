"use client";

import { useMemo } from "react";
import { toast } from "sonner";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type {
  Contact,
  Journal,
  PriceBook,
  SaleOrder,
  Shipment,
  Warehouse,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export type ReminderAction = {
  id: number;
  contactId: number;
  invoiceId: number;
  levelId: number;
  sentAt: string | null;
};

export function useSaleOrderDetail(orgId: string, saleOrderId: string) {
  const numericOrgId = Number(orgId);

  const orderQuery = useOrgQuery<{ order: SaleOrder }>("saleOrder", saleOrderId, (organizationId) =>
    getSwantaraService().saleOrders.get(organizationId, Number(saleOrderId)),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const priceBooksQuery = useOrgListQuery<{ priceBooks: PriceBook[] }, Record<string, never>>(
    "price_books",
    (organizationId) => getSwantaraService().priceBooks.list(organizationId),
  );
  const warehousesQuery = useOrgListQuery<{ warehouses: Warehouse[] }, Record<string, never>>(
    "warehouses",
    (organizationId) => getSwantaraService().inventory.warehouses(organizationId),
  );
  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );
  const shipmentsQuery = useOrgListQuery<{ shipments: Shipment[] }, Record<string, never>>(
    "stockShipments",
    (organizationId) => getSwantaraService().inventory.stockShipments(organizationId),
  );
  const reminderQuery = useOrgListQuery<{ actions: ReminderAction[] }, Record<string, never>>(
    "reminder",
    (organizationId) => getSwantaraService().reminder.list(organizationId),
  );

  const order = orderQuery.data?.order ?? null;
  const contacts = contactsQuery.data?.contacts ?? [];
  const priceBooks = priceBooksQuery.data?.priceBooks ?? [];
  const warehouses = warehousesQuery.data?.warehouses ?? [];
  const journals = journalsQuery.data?.journals ?? [];
  const shipments = shipmentsQuery.data?.shipments ?? [];
  const reminders = reminderQuery.data?.actions ?? [];

  const contactMap = useMemo(
    () => new Map(contacts.map((contact) => [contact.id, contact.displayName || contact.name])),
    [contacts],
  );
  const priceBookMap = useMemo(
    () => new Map(priceBooks.map((priceBook) => [priceBook.id, priceBook.name])),
    [priceBooks],
  );
  const warehouseMap = useMemo(
    () => new Map(warehouses.map((warehouse) => [warehouse.id, warehouse.name])),
    [warehouses],
  );

  const lines = order?.lines ?? [];
  const linkedShipment = order
    ? shipments.find((shipment) => shipment.origin === order.name)
    : null;
  const deliveredUnbilledLines = lines.filter((line) => line.qtyDelivered - line.qtyInvoiced > 0);

  function refreshOrder() {
    void orderQuery.refetch();
  }

  function send() {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.send(numericOrgId, order.id)
      .then(() => {
        toast.success("Quotation sent.");
        refreshOrder();
      })
      .catch(() => toast.error("Could not send the quotation."));
  }

  function confirm() {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.confirm(numericOrgId, order.id)
      .then(() => {
        toast.success("Order confirmed — reservation and shipment created.");
        refreshOrder();
        void shipmentsQuery.refetch();
      })
      .catch(() => toast.error("Could not confirm the order."));
  }

  function cancel() {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.cancel(numericOrgId, order.id)
      .then(() => {
        toast.success("Order cancelled.");
        refreshOrder();
      })
      .catch(() => toast.error("Could not cancel the order."));
  }

  function markDone() {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.done(numericOrgId, order.id)
      .then(() => {
        toast.success("Order marked done.");
        refreshOrder();
      })
      .catch(() => toast.error("Could not mark the order done."));
  }

  function deliver(values: { journalId: string; date: string }) {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.deliver(numericOrgId, order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
      })
      .then(() => {
        toast.success("Delivery posted — COGS and stock updated.");
        refreshOrder();
        void shipmentsQuery.refetch();
      })
      .catch(() => toast.error("Could not post the delivery."));
  }

  function invoice(values: { journalId: string; date: string }) {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.invoice(numericOrgId, order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
      })
      .then(() => {
        toast.success("Invoice created.");
        refreshOrder();
      })
      .catch(() => toast.error("Could not create the invoice."));
  }

  function pay(values: { journalId: string; amount: string; date: string }) {
    if (!order) return;
    void getSwantaraService()
      .saleOrders.pay(numericOrgId, order.id, {
        journalId: Number(values.journalId),
        amount: Number(values.amount),
        date: values.date || null,
      })
      .then(() => {
        toast.success("Payment created and allocated.");
        refreshOrder();
      })
      .catch(() => toast.error("Could not create the payment."));
  }

  function generateReminder() {
    void getSwantaraService()
      .reminder.generate(numericOrgId, { organizationId: numericOrgId, asOf: null })
      .then(() => {
        toast.success("Reminder actions generated.");
        void reminderQuery.refetch();
      })
      .catch(() => toast.error("Could not generate reminders."));
  }

  return {
    order,
    lines,
    linkedShipment,
    deliveredUnbilledLines,
    contactMap,
    priceBookMap,
    warehouseMap,
    journals,
    reminders,
    isLoading: orderQuery.isLoading,
    send,
    confirm,
    cancel,
    markDone,
    deliver,
    invoice,
    pay,
    generateReminder,
  };
}

export type SaleOrderDetailData = ReturnType<typeof useSaleOrderDetail>;
