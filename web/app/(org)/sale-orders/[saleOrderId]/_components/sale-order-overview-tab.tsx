"use client";

import { useTranslations } from "next-intl";
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
  const t = useTranslations("Sales");
  const tCommon = useTranslations("Common");
  const lines = order.lines ?? [];

  function stepLabel(step: string): string {
    try {
      return (t as unknown as (k: string) => string)(`orderStep.${step}`);
    } catch {
      return String(step);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <WorkflowSteps
        steps={orderSteps.map((step) => ({ label: stepLabel(step) }))}
        currentIndex={orderStateIndex(order.state)}
      />
      <div className="flex flex-wrap gap-2">
        {canSend(order.state) ? (
          <Button size="sm" onClick={onSend}>
            {t("actionSend")}
          </Button>
        ) : null}
        {canConfirm(order.state) ? (
          <Button size="sm" onClick={onConfirm}>
            {t("actionConfirm")}
          </Button>
        ) : null}
        {canCancel(order.state) ? (
          <Button size="sm" variant="outline" onClick={onCancel}>
            {tCommon("cancel")}
          </Button>
        ) : null}
        {canDone(order.state) ? (
          <Button size="sm" variant="outline" onClick={onDone}>
            {t("actionMarkDone")}
          </Button>
        ) : null}
        {!canEdit(order.state) ? (
          <span className="text-xs text-muted-foreground self-center">{t("draftOnlyHint")}</span>
        ) : null}
      </div>
      <p className="text-xs text-muted-foreground">{t("confirmHint")}</p>
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("orderHeader")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("tableCustomer")}</span>
              <span>{contactMap.get(order.contactId) ?? `#${order.contactId}`}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("fieldPriceBook")}</span>
              <span>
                {order.priceBookId
                  ? (priceBookMap.get(order.priceBookId) ?? `#${order.priceBookId}`)
                  : "—"}
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
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("fieldOpportunity")}</span>
              <span>{order.crmLeadId ? `#${order.crmLeadId}` : "—"}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("tableOrderDate")}</span>
              <span>{order.orderDate ? formatDate(order.orderDate) : "—"}</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("tableTotal")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("untaxed")}</span>
              <span className="tabular-nums">{formatNumber(order.amountUntaxed)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("tax")}</span>
              <span className="tabular-nums">{formatNumber(order.amountTax)}</span>
            </div>
            <div className="flex justify-between">
              <span>{t("tableTotal")}</span>
              <span className="tabular-nums">{formatNumber(order.amountTotal)}</span>
            </div>
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>{t("tableDelivery")}</span>
              <Badge variant="outline">{String(order.deliveryStatus)}</Badge>
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
          <p className="text-xs text-muted-foreground">{t("linesHint")}</p>
        </CardHeader>
        <CardContent>
          {lines.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("emptyLines")}</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2">#</th>
                    <th className="pb-2">{t("fieldItem")}</th>
                    <th className="pb-2 text-right">{t("ordered")}</th>
                    <th className="pb-2 text-right">{t("delivered")}</th>
                    <th className="pb-2 text-right">{t("invoiced")}</th>
                    <th className="pb-2 text-right">{t("remainingToShip")}</th>
                    <th className="pb-2 text-right">{t("remainingToBill")}</th>
                    <th className="pb-2 text-right">{t("unitPrice")}</th>
                    <th className="pb-2 text-right">{t("fieldDiscountPct")}</th>
                    <th className="pb-2 text-right">{t("subtotal")}</th>
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
