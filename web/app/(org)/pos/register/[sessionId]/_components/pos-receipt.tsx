"use client";

import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { formatMoney } from "@/lib/utils";

export function PosReceipt({
  orderId,
  orderName,
  amountTotal,
  payments,
  onNewOrder,
}: {
  orderId: number;
  orderName: string | null;
  amountTotal: number;
  payments: { method: string; amount: number }[];
  onNewOrder: () => void;
}) {
  return (
    <div className="flex h-[calc(100dvh-4rem)] items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-lg">{"Receipt"}</CardTitle>
          <a href={`/pos/orders/${orderId}`} className="text-sm text-primary hover:underline">
            {orderName ?? `POS-${orderId}`}
          </a>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="text-center">
            <p className="text-xs text-muted-foreground">{"Total"}</p>
            <p className="text-2xl font-bold tabular-nums">{formatMoney(amountTotal)}</p>
          </div>

          <div className="flex flex-col gap-2">
            <p className="text-xs text-muted-foreground">{"Payments"}</p>
            {payments.map((p, i) => (
              <div key={`${p.method}-${i}`} className="flex justify-between text-sm">
                <span className="text-muted-foreground">{p.method}</span>
                <span className="tabular-nums ">{formatMoney(p.amount)}</span>
              </div>
            ))}
          </div>

          <div className="flex flex-col gap-2 pt-2">
            <Button className="w-full" onClick={onNewOrder}>
              {"New order"}
            </Button>
            <Button variant="outline" className="w-full" onClick={() => window.print()}>
              {"Print"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
