"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { SaleOrder, Shipment } from "@/lib/services/swantara";
import { canDeliver } from "../../_components/sale-order-utils";

type SaleOrderDeliveryTabProps = {
  order: SaleOrder;
  linkedShipment: Shipment | null | undefined;
  onDeliver: () => void;
};

export function SaleOrderDeliveryTab({
  order,
  linkedShipment,
  onDeliver,
}: SaleOrderDeliveryTabProps) {
  const t = useTranslations("Sales");
  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("tabDelivery")}</CardTitle>
          <p className="text-sm text-muted-foreground">{t("deliverDescription")}</p>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {linkedShipment ? (
            <div className="flex items-center gap-2">
              <span className="text-sm">{t("linkedShipment")}:</span>
              <a
                href={`/stock/shipments/${linkedShipment.id}`}
                className="text-sm text-primary underline"
              >
                {linkedShipment.name ?? `PK-${linkedShipment.id}`} — {linkedShipment.state}
              </a>
              <Badge variant="outline">{linkedShipment.state}</Badge>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">{t("noShipmentHint")}</p>
          )}
          {canDeliver(order.state) ? (
            <Button size="sm" onClick={onDeliver}>
              {t("deliverTitle")}
            </Button>
          ) : (
            <p className="text-xs text-muted-foreground">{t("confirmToDeliverHint")}</p>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("deliveryStatusTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 text-sm">
            <span>
              {t("tableDelivery")}: <Badge variant="outline">{String(order.deliveryStatus)}</Badge>
            </span>
            <span>
              {t("tableInvoicing")}: <Badge variant="outline">{String(order.invoiceStatus)}</Badge>
            </span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
