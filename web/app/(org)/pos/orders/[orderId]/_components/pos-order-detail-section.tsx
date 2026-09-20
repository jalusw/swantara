"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import { JournalDateDialog } from "@/components/journal-date-dialog";
import { RecordLayout } from "@/components/record-layout";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Journal, PosOrder, PosOrderLine, PosPayment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import {
  canInvoiceOrder,
  canRefundOrder,
  posOrderStateConfig,
} from "../../../_components/pos-utils";

export function PosOrderDetail({ orgId, orderId }: { orgId: string; orderId: string }) {
  const [invoiceOpen, setInvoiceOpen] = useState(false);
  const [refundOpen, setRefundOpen] = useState(false);

  const orderQuery = useOrgQuery<{ order: PosOrder }>("posOrder", orderId, (organizationId) =>
    getSwantaraService().posOrders.get(organizationId, Number(orderId)),
  );
  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );

  const order = orderQuery.data?.order ?? null;
  const journals = journalsQuery.data?.journals ?? [];

  if (orderQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!order) {
    return <p className="text-sm text-muted-foreground">{"Order not found."}</p>;
  }

  const lines: PosOrderLine[] = order.lines ?? [];
  const payments: PosPayment[] = order.payments ?? [];

  function handleInvoice(values: { journalId: string; date: string }) {
    void getSwantaraService()
      .posOrders.invoice(Number(orgId), order!.id, {
        journalId: Number(values.journalId),
        date: values.date,
      })
      .then(({ invoiceId }) => {
        toast.success(`Invoice #${invoiceId} created.`);
        setInvoiceOpen(false);
        void orderQuery.refetch();
      })
      .catch(() => toast.error("Something went wrong. Please try again."));
  }

  function handleRefund(values: { journalId: string; date: string }) {
    void getSwantaraService()
      .posOrders.refund(Number(orgId), order!.id, {
        journalId: Number(values.journalId),
        date: values.date,
      })
      .then(() => {
        toast.success("Order refunded.");
        setRefundOpen(false);
        void orderQuery.refetch();
      })
      .catch(() => toast.error("Something went wrong. Please try again."));
  }

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: "POS Orders", href: "/pos/orders" },
          { label: order.name ?? `POS-${order.id}` },
        ]}
        title={order.name ?? `POS-${order.id}`}
        status={
          <StateBadge
            value={order.state}
            statuses={{
              done: posOrderStateConfig("done"),
              refunded: posOrderStateConfig("refunded"),
            }}
          />
        }
        actions={
          <div className="flex gap-2">
            {canInvoiceOrder(order) ? (
              <Button size="sm" onClick={() => setInvoiceOpen(true)}>
                {"Create invoice"}
              </Button>
            ) : null}
            {canRefundOrder(order) ? (
              <Button size="sm" variant="outline" onClick={() => setRefundOpen(true)}>
                {"Refund"}
              </Button>
            ) : null}
            {order.invoiceId ? (
              <Button size="sm" variant="ghost" asChild>
                <a href={`/accounting/invoices/${order.invoiceId}`}>{"View invoice"}</a>
              </Button>
            ) : null}
          </div>
        }
        tabs={[
          {
            id: "overview",
            label: "Overview",
            content: (
              <div className="flex flex-col gap-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{"Order info"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Session"}</span>
                        <a
                          href={`/pos/sessions/${order.sessionId}`}
                          className=" text-primary hover:underline"
                        >
                          Session-{order.sessionId}
                        </a>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Customer"}</span>
                        <span>{order.contactId ? `#${order.contactId}` : "—"}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Order time"}</span>
                        <span>{order.orderTime ? formatDate(order.orderTime) : "—"}</span>
                      </div>
                      {order.invoiceId ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{"Invoice"}</span>
                          <a
                            href={`/accounting/invoices/${order.invoiceId}`}
                            className=" text-primary hover:underline"
                          >
                            #{order.invoiceId}
                          </a>
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
                        <span className="text-muted-foreground">{"Subtotal"}</span>
                        <span className="tabular-nums">
                          {formatMoney(order.amountTotal - order.amountTax)}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Tax"}</span>
                        <span className="tabular-nums">{formatMoney(order.amountTax)}</span>
                      </div>
                      <div className="flex justify-between ">
                        <span>{"Total"}</span>
                        <span className="tabular-nums">{formatMoney(order.amountTotal)}</span>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </div>
            ),
          },
          {
            id: "lines",
            label: "Lines",
            content: (
              <Card>
                <CardContent className="pt-6">
                  {lines.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{"No lines in this order."}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-left text-muted-foreground">
                            <th className="pb-2 ">#</th>
                            <th className="pb-2 ">{"Item"}</th>
                            <th className="pb-2 text-right">{"Qty"}</th>
                            <th className="pb-2 text-right">{"Unit price"}</th>
                            <th className="pb-2 text-right">{"Discount"}</th>
                            <th className="pb-2 text-right">{"Subtotal"}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {lines.map((line: PosOrderLine, idx: number) => (
                            <tr key={line.id} className="border-b last:border-0">
                              <td className="py-2">{idx + 1}</td>
                              <td className="py-2">{line.itemId ?? "—"}</td>
                              <td className="py-2 text-right tabular-nums">
                                {formatNumber(line.qty)}
                              </td>
                              <td className="py-2 text-right tabular-nums">
                                {formatMoney(line.unitPrice)}
                              </td>
                              <td className="py-2 text-right tabular-nums">
                                {line.discountPct > 0 ? `${line.discountPct}%` : "—"}
                              </td>
                              <td className="py-2 text-right tabular-nums ">
                                {formatMoney(line.priceSubtotal)}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </CardContent>
              </Card>
            ),
          },
          {
            id: "payments",
            label: "Payments",
            content: (
              <Card>
                <CardContent className="pt-6">
                  {payments.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{"No payments recorded."}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-left text-muted-foreground">
                            <th className="pb-2 ">{"Method"}</th>
                            <th className="pb-2 text-right">{"Amount"}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {payments.map((payment: PosPayment) => (
                            <tr key={payment.id} className="border-b last:border-0">
                              <td className="py-2">
                                <Badge variant="outline">
                                  {humanizeKey(String(payment.method))}
                                </Badge>
                              </td>
                              <td className="py-2 text-right tabular-nums ">
                                {formatMoney(payment.amount)}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                        <tfoot>
                          <tr className="border-t ">
                            <td className="py-2">{"Total"}</td>
                            <td className="py-2 text-right tabular-nums">
                              {formatMoney(payments.reduce((sum, p) => sum + p.amount, 0))}
                            </td>
                          </tr>
                        </tfoot>
                      </table>
                    </div>
                  )}
                </CardContent>
              </Card>
            ),
          },
        ]}
      />
      {invoiceOpen ? (
        <JournalDateDialog
          title={"Create invoice"}
          description={"Generate a customer invoice for this POS order."}
          confirmLabel={"Create invoice"}
          cancelLabel={"Cancel"}
          open={invoiceOpen}
          onOpenChange={setInvoiceOpen}
          onConfirm={(journalId, date) => handleInvoice({ journalId, date })}
          journals={journals}
        />
      ) : null}
      {refundOpen ? (
        <JournalDateDialog
          title={"Refund"}
          description={"Issue a refund for this POS order. This will create a credit note."}
          confirmLabel={"Refund"}
          cancelLabel={"Cancel"}
          open={refundOpen}
          onOpenChange={setRefundOpen}
          onConfirm={(journalId, date) => handleRefund({ journalId, date })}
          journals={journals}
        />
      ) : null}
    </>
  );
}
