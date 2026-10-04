"use client";

import { useTranslations } from "next-intl";
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

function getPoStatuses(label?: (state: string) => string) {
  const text = (state: string, fallback: string) => (label ? label(state) : fallback);
  return {
    draft: { label: text("draft", "Draft"), tone: "neutral" as const },
    sent: { label: text("sent", "Terkirim"), tone: "info" as const },
    confirmed: { label: text("confirmed", "Dikonfirmasi"), tone: "warning" as const },
    done: { label: text("done", "Selesai"), tone: "success" as const },
    cancelled: { label: text("cancelled", "Dibatalkan"), tone: "danger" as const },
  };
}

export function PurchaseOrderDetail({ orgId, orderId }: { orgId: string; orderId: string }) {
  const t = useTranslations("Purchases");
  const tCommon = useTranslations("Common");
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
    return <p className="text-sm text-muted-foreground">{t("purchaseOrderNotFound")}</p>;
  }

  function poStateLabel(state: string): string {
    try {
      return (t as unknown as (k: string) => string)(`poState.${state}`);
    } catch {
      return state;
    }
  }

  function poStepLabel(step: string): string {
    try {
      return (t as unknown as (k: string) => string)(`poStep.${step}`);
    } catch {
      return String(step);
    }
  }

  const lines = order.lines ?? [];
  const linkedShipment = shipments.find((p) => p.origin === order.name);

  function handleConfirm() {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.confirm(Number(orgId), order.id)
      .then(() => {
        toast.success(t("orderConfirmed"));
        void orderQuery.refetch();
        void shipmentsQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleCancel() {
    if (!order) return;
    void getSwantaraService()
      .purchaseOrders.cancel(Number(orgId), order.id)
      .then(() => {
        toast.success(t("orderCancelled"));
        void orderQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleReceive(values: { journalId: string; date: string }) {
    if (!order) return;
    return getSwantaraService()
      .purchaseOrders.receive(Number(orgId), order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
      })
      .then(() => {
        toast.success(t("goodsReceived"));
        setActiveDialog(null);
        void orderQuery.refetch();
        void shipmentsQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleBill(values: { journalId: string; date: string; override: boolean }) {
    if (!order) return;
    return getSwantaraService()
      .purchaseOrders.vendorBill(Number(orgId), order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
        override: values.override,
      })
      .then(() => {
        toast.success(t("supplierBillCreated"));
        setActiveDialog(null);
        void orderQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handlePay(values: { journalId: string; date: string }) {
    if (!order) return;
    return getSwantaraService()
      .purchaseOrders.pay(Number(orgId), order.id, {
        journalId: Number(values.journalId),
        date: values.date || null,
      })
      .then(() => {
        toast.success(t("paymentRecorded"));
        setActiveDialog(null);
        void orderQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  const billableLines = lines.filter((line) => line.qtyOrdered - line.qtyBilled > 0);

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: t("allPurchaseOrders"), href: "/purchases" },
          { label: order.name ?? `PO-${order.id}` },
        ]}
        title={order.name ?? `PO-${order.id}`}
        status={<StateBadge value={order.state} statuses={getPoStatuses(poStateLabel)} />}
        tabs={[
          {
            id: "overview",
            label: t("tabOverview"),
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps
                  steps={poSteps.map((s) => ({ label: poStepLabel(s) }))}
                  currentIndex={poStateIndex(order.state)}
                />
                <div className="flex flex-wrap gap-2">
                  {canConfirm(order.state) ? (
                    <Button size="sm" onClick={handleConfirm}>
                      {t("actionConfirm")}
                    </Button>
                  ) : null}
                  {canCancel(order.state) ? (
                    <Button size="sm" variant="outline" onClick={handleCancel}>
                      {tCommon("cancel")}
                    </Button>
                  ) : null}
                  {!canEdit(order.state) ? (
                    <span className="text-xs text-muted-foreground self-center">
                      {t("draftOnlyHint")}
                    </span>
                  ) : null}
                </div>
                <p className="text-xs text-muted-foreground">{t("confirmPurchaseHint")}</p>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{t("orderHeader")}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableSupplier")}</span>
                        <span className="">
                          {contactMap.get(order.supplierId) ?? `#${order.supplierId}`}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("fieldWarehouse")}</span>
                        <span>
                          {order.warehouseId
                            ? (warehouseMap.get(order.warehouseId) ?? `#${order.warehouseId}`)
                            : "—"}
                        </span>
                      </div>
                      {order.vendorRef ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{t("fieldVendorRef")}</span>
                          <span>{order.vendorRef}</span>
                        </div>
                      ) : null}
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableOrderDate")}</span>
                        <span>{order.orderDate ? formatDate(order.orderDate) : "—"}</span>
                      </div>
                      {order.expectedDate ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{t("fieldExpectedDate")}</span>
                          <span>{formatDate(order.expectedDate)}</span>
                        </div>
                      ) : null}
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{t("tableTotal")}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("untaxed")}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountUntaxed, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tax")}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountTax, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </div>
                      <div className="flex justify-between ">
                        <span>{t("tableTotal")}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountTotal, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </div>
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>{t("tableReceipt")}</span>
                        <Badge variant="outline">{String(order.receiptStatus)}</Badge>
                      </div>
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>{t("tableInvoicing")}</span>
                        <Badge variant="outline">{String(order.invoiceStatus)}</Badge>
                      </div>
                    </CardContent>
                  </Card>
                </div>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{t("orderLines")}</CardTitle>
                    <p className="text-xs text-muted-foreground">{t("poLinesHint")}</p>
                  </CardHeader>
                  <CardContent>
                    {lines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">{t("emptyLines")}</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b text-left text-muted-foreground">
                              <th className="pb-2 ">#</th>
                              <th className="pb-2 ">{t("fieldItem")}</th>
                              <th className="pb-2 text-right">{t("ordered")}</th>
                              <th className="pb-2 text-right">{t("received")}</th>
                              <th className="pb-2 text-right">{t("billed")}</th>
                              <th className="pb-2 text-right">{t("toReceive")}</th>
                              <th className="pb-2 text-right">{t("toBill")}</th>
                              <th className="pb-2 text-right">{t("unitPrice")}</th>
                              <th className="pb-2 text-right">{t("subtotal")}</th>
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
            label: t("tabReceipt"),
            content: (
              <div className="flex flex-col gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{t("tabReceipt")}</CardTitle>
                    <p className="text-sm text-muted-foreground">{t("receiptDescription")}</p>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {linkedShipment ? (
                      <div className="flex items-center gap-2">
                        <span className="text-sm">{t("linkedShipment")}:</span>
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
                      <p className="text-sm text-muted-foreground">{t("noShipmentHint")}</p>
                    )}
                    {canReceive(order.state) ? (
                      <Button size="sm" onClick={() => setActiveDialog("receive")}>
                        {t("actionReceive")}
                      </Button>
                    ) : (
                      <p className="text-xs text-muted-foreground">{t("confirmToReceiveHint")}</p>
                    )}
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{t("receiptStatusTitle")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-sm">
                      <span>
                        {t("tableReceipt")}:{" "}
                        <Badge variant="outline">{String(order.receiptStatus)}</Badge>
                      </span>
                    </div>
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            id: "invoicing",
            label: t("tabInvoicing"),
            content: (
              <div className="flex flex-col gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{t("tabInvoicing")}</CardTitle>
                    <p className="text-sm text-muted-foreground">{t("poBillingDescription")}</p>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {billableLines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">{t("noBillableLines")}</p>
                    ) : (
                      <>
                        <p className="text-sm">
                          {t("linesToBill")}: {billableLines.length}
                        </p>
                        <div className="overflow-x-auto">
                          <table className="w-full text-sm">
                            <thead>
                              <tr className="border-b text-muted-foreground">
                                <th className="pb-2 text-left">{t("fieldItem")}</th>
                                <th className="pb-2 text-right">{t("orderedNotBilled")}</th>
                                <th className="pb-2 text-right">{t("unitPrice")}</th>
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
                      {t("createSupplierBill")}
                    </Button>
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            id: "payments",
            label: t("tabPayments"),
            content: (
              <div className="flex flex-col gap-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{t("tabPayments")}</CardTitle>
                    <p className="text-sm text-muted-foreground">{t("poPaymentsDescription")}</p>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    <div className="flex gap-4 text-sm">
                      <span>
                        {t("tableTotal")}:{" "}
                        <span className=" tabular-nums">
                          {formatMoney(order.amountTotal, { currency: DEFAULT_CURRENCY })}
                        </span>
                      </span>
                      <span className="text-muted-foreground">{t("residualHint")}</span>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => setActiveDialog("pay")}
                      disabled={!canPay(order.state)}
                    >
                      {t("createPayment")}
                    </Button>
                  </CardContent>
                </Card>
              </div>
            ),
          },
        ]}
      />
      <JournalDateDialog
        title={t("actionReceive")}
        description={t("receiptDescription")}
        confirmLabel={t("actionReceive")}
        cancelLabel={tCommon("cancel")}
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
        title={t("createPayment")}
        confirmLabel={t("createPayment")}
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
  const t = useTranslations("Purchases");
  const tCommon = useTranslations("Common");
  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());
  const [override, setOverride] = useState(false);
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("createSupplierBill")}</DialogTitle>
          <DialogDescription>{t("poBillingDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldJournal")}</span>
            <Select value={journalId} onValueChange={(value) => setJournalId(value ?? "")}>
              <SelectTrigger aria-label={t("fieldJournal")}>
                <SelectValue placeholder={t("fieldJournal")} />
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
            <span className="text-sm">{t("tableOrderDate")}</span>
            <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={override}
              onChange={(e) => setOverride(e.target.checked)}
              className="size-4"
            />
            {t("overrideHint")}
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={() => onSubmit({ journalId, date, override })} disabled={!journalId}>
            {t("createSupplierBill")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
