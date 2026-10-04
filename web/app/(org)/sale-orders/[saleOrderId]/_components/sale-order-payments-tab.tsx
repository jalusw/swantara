"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { SaleOrder } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

type SaleOrderPaymentsTabProps = {
  order: SaleOrder;
  onPay: () => void;
};

export function SaleOrderPaymentsTab({ order, onPay }: SaleOrderPaymentsTabProps) {
  const t = useTranslations("Sales");
  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("tabPayments")}</CardTitle>
          <p className="text-sm text-muted-foreground">{t("payDescription")}</p>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <div className="flex gap-4 text-sm">
            <span>
              {t("tableTotal")}:{" "}
              <span className="tabular-nums">{formatNumber(order.amountTotal)}</span>
            </span>
            <span className="text-muted-foreground">{t("residualHint")}</span>
          </div>
          <Button size="sm" onClick={onPay}>
            {t("payTitle")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
