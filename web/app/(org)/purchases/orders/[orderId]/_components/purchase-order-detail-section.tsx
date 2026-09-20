"use client";

import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { JournalDateDialog } from "@/components/journal-date-dialog";
import { RecordLayout } from "@/components/record-layout";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { StateBadge } from "@/components/state-badge";
import { WorkflowSteps } from "@/components/workflow-steps";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type {
  Contact,
  Journal,
  PurchaseOrder,
  PurchaseOrderLine,
  Shipment,
  Warehouse,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney, formatNumber, getLocalDateString } from "@/lib/utils";
import {
  canCancel,
  canConfirm,
  canEdit,
  canPay,
  canReceive,
  remainingToBill,
  remainingToReceive,
} from "../../_components/purchase-order-utils";

const poSteps = ["draft", "sent", "confirmed", "done"] as const;

function poStateIndex(state: PurchaseOrder["state"]): number {
  const idx = poSteps.indexOf(state as (typeof poSteps)[number]);
  return idx >= 0 ? idx : 0;
}

function getPoStatuses() {
  return {
    draft: { label: "Draft", tone: "neutral" as const },
    sent: { label: "Sent", tone: "info" as const },
    confirmed: { label: "Confirmed", tone: "warning" as const },
    done: { label: "Done", tone: "success" as const },
    cancelled: { label: "Cancelled", tone: "danger" as const },
  };
}

export function PurchaseOrderDetail({ orgId, orderId }: { orgId: string; orderId: string }) {
  const [activeDialog, setActiveDialog] = useState<"receive" | "bill" | "pay" | null>(null);

  const orderQuery = useOrgQuery<{ order: PurchaseOrder }>(
    "purchaseOrder",
    orderId,
    (organizationId) => getSwantaraService().purchaseOrders.get(organizationId, Number(orderId)),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
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

  const order = orderQuery.data?.order ?? null;
  const contacts = contactsQuery.data?.contacts ?? [];
  const warehouses = warehousesQuery.data?.warehouses ?? [];
  const journals = journalsQuery.data?.journals ?? [];
  const shipments = shipmentsQuery.data?.shipments ?? [];

  const contactMap = useMemo(
    () => new Map(contacts.map((p) => [p.id, p.displayName || p.name])),
    [contacts],
  );
  const warehouseMap = useMemo(() => new Map(warehouses.map((w) => [w.id, w.name])), [warehouses]);

  if (orderQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!order) {
    return <p className="text-sm text-muted-foreground">{"Purchase order not found."}</p>;
  }

  const lines = order.lines ?? [];
  const linkedShipment = shipments.find((p) => p.origin === order.name);

  function handleConfirm() {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.confirm(Number(orgId), order.id)
      .then(() => {
        toast.success("Order confirmed — incoming shipment created.");
        void orderQuery.refetch();
        void shipmentsQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleCancel() {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.cancel(Number(orgId), order.id)
      .then(() => {
        toast.success("Order cancelled.");
        void orderQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleReceive(values: { journalId: string; date: string }) {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.receive(Number(orgId), order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
      })
      .then(() => {
        toast.success("Goods received.");
        setActiveDialog(null);
        void orderQuery.refetch();
        void shipmentsQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleBill(values: { journalId: string; date: string; override: boolean }) {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.vendorBill(Number(orgId), order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
        override: values.override,
      })
      .then(() => {
        toast.success("Supplier bill created.");
        setActiveDialog(null);
        void orderQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handlePay(values: { journalId: string; date: string }) {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.pay(Number(orgId), order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
      })
      .then(() => {
        toast.success("Payment recorded.");
        setActiveDialog(null);
        void orderQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  const billableLines = lines.filter((line) => line.qtyOrdered - line.qtyBilled > 0);

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: "All purchase orders", href: "/purchases" },
          { label: order.name ?? `PO-${order.id}` },
        ]}
        title={order.name ?? `PO-${order.id}`}
        status={<StateBadge value={order.state} statuses={getPoStatuses()} />}
        tabs={[
          {
            id: "overview",
            label: "Overview",
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps
                  steps={poSteps.map((s) => ({ label: String(s) }))}
                  currentIndex={poStateIndex(order.state)}
                />
                <div className="flex flex-wrap gap-2">
                  {canConfirm(order.state) ? (
                    <Button size="sm" onClick={handleConfirm}>
                      {"Confirm"}
                    </Button>
                  ) : null}
                  {canCancel(order.state) ? (
                    <Button size="sm" variant="outline" onClick={handleCancel}>
                      {"Cancel"}
                    </Button>
                  ) : null}
                  {!canEdit(order.state) ? (
                    <span className="text-xs text-muted-foreground self-center">
                      {"Only draft orders can be edited."}
                    </span>
                  ) : null}
                </div>
                <p className="text-xs text-muted-foreground">
                  {"Confirm creates an incoming shipment for receiving."}
                </p>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{"Order header"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Supplier"}</span>
                        <span className="">
                          {contactMap.get(order.supplierId) ?? `#${order.supplierId}`}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Warehouse"}</span>
                        <span>
                          {order.warehouseId
                            ? (warehouseMap.get(order.warehouseId) ?? `#${order.warehouseId}`)
                            : "—"}
                        </span>
                      </div>
                      {order.vendorRef ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{"Supplier reference"}</span>
                          <span>{order.vendorRef}</span>
                        </div>
                      ) : null}
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Order date"}</span>
                        <span>{order.orderDate ? formatDate(order.orderDate) : "—"}</span>
                      </div>
                      {order.expectedDate ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{"Expected date"}</span>
                          <span>{formatDate(order.expectedDate)}</span>
                        </div>
                      ) : null}
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{"Totals"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Untaxed"}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountUntaxed, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Tax"}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountTax, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </div>
                      <div className="flex justify-between ">
                        <span>{"Total"}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountTotal, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </div>
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>{"Receipt"}</span>
                        <Badge variant="outline">{String(order.receiptStatus)}</Badge>
                      </div>
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>{"Invoicing"}</span>
                        <Badge variant="outline">{String(order.invoiceStatus)}</Badge>
                      </div>
                    </CardContent>
                  </Card>
                </div>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{"Lines"}</CardTitle>
                    <p className="text-xs text-muted-foreground">
                      {"Derived from ordered/received/billed"}
                    </p>
                  </CardHeader>
                  <CardContent>
                    {lines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">{"No lines."}</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b text-left text-muted-foreground">
                              <th className="pb-2 ">#</th>
                              <th className="pb-2 ">{"Item"}</th>
                              <th className="pb-2 text-right">{"Ordered"}</th>
                              <th className="pb-2 text-right">{"Received"}</th>
                              <th className="pb-2 text-right">{"Billed"}</th>
                              <th className="pb-2 text-right">{"To receive"}</th>
                              <th className="pb-2 text-right">{"To bill"}</th>
                              <th className="pb-2 text-right">{"Unit price"}</th>
                              <th className="pb-2 text-right">{"Subtotal"}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {lines.map((line: PurchaseOrderLine, idx: number) => (
                              <tr key={line.id} className="border-b last:border-0">
                                <td className="py-2">{idx + 1}</td>
                                <td className="py-2">{line.itemId ?? "—"}</td>
                                <td className="py-2 text-right tabular-nums">
                                  {formatNumber(line.qtyOrdered)}
                                </td>
                                <td className="py-2 text-right tabular-nums">
                                  {formatNumber(line.qtyReceived)}
                                </td>
                                <td className="py-2 text-right tabular-nums">
                                  {formatNumber(line.qtyBilled)}
                                </td>
                                <td className="py-2 text-right tabular-nums ">
                                  {formatMoney(remainingToReceive(line), {
                                    currency: DEFAULT_CURRENCY,
                                  })}
                                </td>
                                <td className="py-2 text-right tabular-nums ">
                                  {formatMoney(remainingToBill(line), {
                                    currency: DEFAULT_CURRENCY,
                                  })}
                                </td>
                                <td className="py-2 text-right tabular-nums">
                                  {formatMoney(line.unitPrice, { currency: DEFAULT_CURRENCY })}
                                </td>
                                <td className="py-2 text-right tabular-nums">
                                  {formatMoney(line.priceSubtotal, { currency: DEFAULT_CURRENCY })}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            id: "receipt",
            label: "Receipt",
            content: (
              <div className="flex flex-col gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{"Receipt"}</CardTitle>
                    <p className="text-sm text-muted-foreground">
                      {"Receive goods against this purchase order."}
                    </p>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {linkedShipment ? (
                      <div className="flex items-center gap-2">
                        <span className="text-sm">{"Linked shipment"}:</span>
                        <a
                          href={`/stock/shipments/${linkedShipment.id}`}
                          className="text-sm text-primary underline"
                        >
                          {linkedShipment.name ?? `PK-${linkedShipment.id}`} —{" "}
                          {linkedShipment.state}
                        </a>
                        <Badge variant="outline">{linkedShipment.state}</Badge>
                      </div>
                    ) : (
                      <p className="text-sm text-muted-foreground">
                        {"No shipment yet — confirm the order to create one."}
                      </p>
                    )}
                    {canReceive(order.state) ? (
                      <Button size="sm" onClick={() => setActiveDialog("receive")}>
                        {"Receive"}
                      </Button>
                    ) : (
                      <p className="text-xs text-muted-foreground">
                        {"Confirm the order to enable receiving."}
                      </p>
                    )}
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{"Receipt status"}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-sm">
                      <span>
                        {"Receipt"}: <Badge variant="outline">{String(order.receiptStatus)}</Badge>
                      </span>
                    </div>
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            id: "invoicing",
            label: "Invoicing",
            content: (
              <div className="flex flex-col gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{"Invoicing"}</CardTitle>
                    <p className="text-sm text-muted-foreground">
                      {
                        "Create supplier bills for received-not-billed lines. Shows 3-way match status."
                      }
                    </p>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {billableLines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">
                        {"No received-not-billed lines to bill."}
                      </p>
                    ) : (
                      <>
                        <p className="text-sm">
                          {"Lines to bill"}: {billableLines.length}
                        </p>
                        <div className="overflow-x-auto">
                          <table className="w-full text-sm">
                            <thead>
                              <tr className="border-b text-muted-foreground">
                                <th className="pb-2 text-left">{"Item"}</th>
                                <th className="pb-2 text-right">{"Ordered not billed"}</th>
                                <th className="pb-2 text-right">{"Unit price"}</th>
                              </tr>
                            </thead>
                            <tbody>
                              {billableLines.map((line: PurchaseOrderLine) => (
                                <tr key={line.id} className="border-b">
                                  <td className="py-2">{line.itemId}</td>
                                  <td className="py-2 text-right tabular-nums">
                                    {formatMoney(line.qtyOrdered - line.qtyBilled, {
                                      currency: DEFAULT_CURRENCY,
                                    })}
                                  </td>
                                  <td className="py-2 text-right tabular-nums">
                                    {formatMoney(line.unitPrice, { currency: DEFAULT_CURRENCY })}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </>
                    )}
                    <Button
                      size="sm"
                      onClick={() => setActiveDialog("bill")}
                      disabled={billableLines.length === 0}
                    >
                      {"Create supplier bill"}
                    </Button>
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            id: "payments",
            label: "Payments",
            content: (
              <div className="flex flex-col gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{"Payments"}</CardTitle>
                    <p className="text-sm text-muted-foreground">
                      {"Record outbound payments against supplier bills."}
                    </p>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    <div className="flex gap-4 text-sm">
                      <span>
                        {"Total"}:{" "}
                        <span className=" tabular-nums">
                          {formatMoney(order.amountTotal, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </span>
                      <span className="text-muted-foreground">
                        {"Residual updates after allocation"}
                      </span>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => setActiveDialog("pay")}
                      disabled={!canPay(order.state)}
                    >
                      {"Make payment"}
                    </Button>
                  </CardContent>
                </Card>
              </div>
            ),
          },
        ]}
      />
      <JournalDateDialog
        title={"Receive"}
        description={"Receive goods against this purchase order."}
        confirmLabel={"Receive"}
        cancelLabel={"Cancel"}
        open={activeDialog === "receive"}
        onOpenChange={(open) => setActiveDialog(open ? "receive" : null)}
        onConfirm={(journalId, date) => handleReceive({ journalId, date })}
        journals={journals}
      />
      <BillDialog
        open={activeDialog === "bill"}
        onOpenChange={(open) => setActiveDialog(open ? "bill" : null)}
        journals={journals}
        onSubmit={handleBill}
      />
      <JournalDateDialog
        title={"Make payment"}
        confirmLabel={"Make payment"}
        journals={journals}
        open={activeDialog === "pay"}
        onOpenChange={(open) => setActiveDialog(open ? "pay" : null)}
        onConfirm={(journalId, date) => handlePay({ journalId, date })}
      />
    </>
  );
}

function BillDialog({
  open,
  onOpenChange,
  journals,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  journals: Journal[];
  onSubmit: (values: { journalId: string; date: string; override: boolean }) => void;
}) {
  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());
  const [override, setOverride] = useState(false);
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Create supplier bill"}</DialogTitle>
          <DialogDescription>
            {"Create supplier bills for received-not-billed lines. Shows 3-way match status."}
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Journal"}</span>
            <Select value={journalId} onValueChange={(value) => setJournalId(value ?? "")}>
              <SelectTrigger aria-label={"Journal"}>
                <SelectValue placeholder={"Journal"} />
              </SelectTrigger>
              <SelectContent>
                {journals.map((j) => (
                  <SelectItem key={j.id} value={String(j.id)}>
                    {j.name} — {j.type}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Date"}</span>
            <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={override}
              onChange={(e) => setOverride(e.target.checked)}
              className="size-4"
            />
            {"Override (bill more than received)"}
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={() => onSubmit({ journalId, date, override })} disabled={!journalId}>
            {"Create supplier bill"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
