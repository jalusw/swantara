"use client";

import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { SaleOrder } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

type SaleOrderPaymentsTabProps = {
  order: SaleOrder;
  onPay: () => void;
};

export function SaleOrderPaymentsTab({ order, onPay }: SaleOrderPaymentsTabProps) {
  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{"Payments"}</CardTitle>
          <p className="text-sm text-muted-foreground">
            {"Capture payment and allocate to invoices. Shows payment_state and residual."}
          </p>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <div className="flex gap-4 text-sm">
            <span>
              {"Total"}: <span className="tabular-nums">{formatNumber(order.amountTotal)}</span>
            </span>
            <span className="text-muted-foreground">{"Residual updates after allocation"}</span>
          </div>
          <Button size="sm" onClick={onPay}>
            {"Collect payment"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
