"use client";

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
  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{"Delivery"}</CardTitle>
          <p className="text-sm text-muted-foreground">
            {"Pick, pack and ship the order. Validates moves and posts COGS."}
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
                {linkedShipment.name ?? `PK-${linkedShipment.id}`} — {linkedShipment.state}
              </a>
              <Badge variant="outline">{linkedShipment.state}</Badge>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              {"No shipment yet — confirm the order to create one."}
            </p>
          )}
          {canDeliver(order.state) ? (
            <Button size="sm" onClick={onDeliver}>
              {"Ship / Deliver"}
            </Button>
          ) : (
            <p className="text-xs text-muted-foreground">
              {"Confirm the order to enable delivery."}
            </p>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{"Delivery status"}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 text-sm">
            <span>
              {"Delivery"}: <Badge variant="outline">{String(order.deliveryStatus)}</Badge>
            </span>
            <span>
              {"Invoicing"}: <Badge variant="outline">{String(order.invoiceStatus)}</Badge>
            </span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
