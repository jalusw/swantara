"use client";

import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { WorkflowSteps } from "@/components/workflow-steps";
import type { SaleOrder } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import {
  canCancel,
  canConfirm,
  canDone,
  canEdit,
  canSend,
  remainingToBill,
  remainingToShip,
} from "../../_components/sale-order-utils";
import { orderStateIndex, orderSteps } from "./sale-order-status";

type SaleOrderOverviewTabProps = {
  order: SaleOrder;
  contactMap: Map<number, string>;
  priceBookMap: Map<number, string>;
  warehouseMap: Map<number, string>;
  onSend: () => void;
  onConfirm: () => void;
  onCancel: () => void;
  onDone: () => void;
};

export function SaleOrderOverviewTab({
  order,
  contactMap,
  priceBookMap,
  warehouseMap,
  onSend,
  onConfirm,
  onCancel,
  onDone,
}: SaleOrderOverviewTabProps) {
  const lines = order.lines ?? [];

  return (
    <div className="flex flex-col gap-4">
      <WorkflowSteps
        steps={orderSteps.map((step) => ({ label: String(step) }))}
        currentIndex={orderStateIndex(order.state)}
      />
      <div className="flex flex-wrap gap-2">
        {canSend(order.state) ? (
          <Button size="sm" onClick={onSend}>
            {"Send"}
          </Button>
        ) : null}
        {canConfirm(order.state) ? (
          <Button size="sm" onClick={onConfirm}>
            {"Confirm"}
          </Button>
        ) : null}
        {canCancel(order.state) ? (
          <Button size="sm" variant="outline" onClick={onCancel}>
            {"Cancel"}
          </Button>
        ) : null}
        {canDone(order.state) ? (
          <Button size="sm" variant="outline" onClick={onDone}>
            {"Mark done"}
          </Button>
        ) : null}
        {!canEdit(order.state) ? (
          <span className="text-xs text-muted-foreground self-center">
            {"Only draft orders can be edited."}
          </span>
        ) : null}
      </div>
      <p className="text-xs text-muted-foreground">
        {"Confirm reserves stock and creates an outgoing shipment."}
      </p>
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{"Order header"}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Customer"}</span>
              <span>{contactMap.get(order.contactId) ?? `#${order.contactId}`}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"PriceBook"}</span>
              <span>
                {order.priceBookId
                  ? (priceBookMap.get(order.priceBookId) ?? `#${order.priceBookId}`)
                  : "—"}
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
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Opportunity"}</span>
              <span>{order.crmLeadId ? `#${order.crmLeadId}` : "—"}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Order date"}</span>
              <span>{order.orderDate ? formatDate(order.orderDate) : "—"}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{"Totals"}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Untaxed"}</span>
              <span className="tabular-nums">{formatNumber(order.amountUntaxed)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Tax"}</span>
              <span className="tabular-nums">{formatNumber(order.amountTax)}</span>
            </div>
            <div className="flex justify-between">
              <span>{"Total"}</span>
              <span className="tabular-nums">{formatNumber(order.amountTotal)}</span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>{"Delivery"}</span>
              <Badge variant="outline">{String(order.deliveryStatus)}</Badge>
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
            {"Derived from ordered/delivered/invoiced"}
          </p>
        </CardHeader>
        <CardContent>
          {lines.length === 0 ? (
            <p className="text-sm text-muted-foreground">{"No lines"}</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2">#</th>
                    <th className="pb-2">{"Item"}</th>
                    <th className="pb-2 text-right">{"Ordered"}</th>
                    <th className="pb-2 text-right">{"Delivered"}</th>
                    <th className="pb-2 text-right">{"Invoiced"}</th>
                    <th className="pb-2 text-right">{"Remaining to ship"}</th>
                    <th className="pb-2 text-right">{"Remaining to bill"}</th>
                    <th className="pb-2 text-right">{"Unit price"}</th>
                    <th className="pb-2 text-right">{"Discount"}</th>
                    <th className="pb-2 text-right">{"Subtotal"}</th>
                  </tr>
                </thead>
                <tbody>
                  {lines.map((line, idx) => (
                    <tr key={line.id} className="border-b last:border-0">
                      <td className="py-2">{idx + 1}</td>
                      <td className="py-2">{line.itemId ?? "—"}</td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(line.qtyOrdered)}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(line.qtyDelivered)}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(line.qtyInvoiced)}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(remainingToShip(line))}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(remainingToBill(line))}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(line.unitPrice)}
                      </td>
                      <td className="py-2 text-right tabular-nums">{line.discountPct}%</td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(line.priceSubtotal)}
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
  );
}
