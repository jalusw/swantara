"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import { RecordLayout } from "@/components/record-layout";
import { WorkflowSteps } from "@/components/workflow-steps";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Shipment, StockMovement } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

const shipmentSteps = ["draft", "waiting", "confirmed", "assigned", "done"] as const;

function shipmentStateIndex(state: Shipment["state"]) {
  const idx = shipmentSteps.indexOf(state as (typeof shipmentSteps)[number]);
  return idx >= 0 ? idx : 0;
}

function stateBadgeVariant(state: Shipment["state"]) {
  switch (state) {
    case "done":
      return "default" as const;
    case "cancelled":
      return "outline" as const;
    case "draft":
      return "secondary" as const;
    default:
      return "secondary" as const;
  }
}

function moveStateBadgeVariant(state: StockMovement["state"]) {
  switch (state) {
    case "done":
      return "default" as const;
    case "cancelled":
      return "outline" as const;
    case "draft":
      return "secondary" as const;
    default:
      return "secondary" as const;
  }
}

export function ShipmentDetail({ orgId, shipmentId }: { orgId: string; shipmentId: string }) {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Stock");
  const [processing, setProcessing] = useState(false);

  const shipmentQuery = useOrgQuery<{ shipment: Shipment }>(
    "stockShipment",
    shipmentId,
    (organizationId) =>
      getSwantaraService().inventory.stockShipment(organizationId, Number(shipmentId)),
  );

  const movementsQuery = useOrgListQuery<{ movements: StockMovement[] }, Record<string, never>>(
    "stockMovements",
    (organizationId) => getSwantaraService().inventory.stockMovements(organizationId),
  );

  const shipment = shipmentQuery.data?.shipment;
  const movements = (movementsQuery.data?.movements ?? []).filter(
    (m) => String(m.shipmentId) === shipmentId,
  );

  if (shipmentQuery.isLoading || movementsQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!shipment) {
    return <p className="text-sm text-muted-foreground">{t("shipmentNotFound")}</p>;
  }

  function handleValidate() {
    if (!shipment) return;
    setProcessing(true);
    void getSwantaraService()
      .inventory.stockShipment(Number(orgId), shipment.id)
      .then(() => {
        void shipmentQuery.refetch();
        void movementsQuery.refetch();
      })
      .catch(() => toast.error(t("toastFailed")))
      .finally(() => setProcessing(false));
  }

  function handleDone() {
    if (!shipment) return;
    setProcessing(true);
    void getSwantaraService()
      .inventory.stockShipment(Number(orgId), shipment.id)
      .then(() => {
        void shipmentQuery.refetch();
        void movementsQuery.refetch();
      })
      .catch(() => toast.error(t("toastFailed")))
      .finally(() => setProcessing(false));
  }

  const canValidate = shipment.state === "draft" || shipment.state === "confirmed";
  const canDone = shipment.state === "assigned";

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("allShipments"), href: "/stock/shipments" },
        { label: shipment.name ?? `PK-${shipment.id}` },
      ]}
      title={shipment.name ?? `PK-${shipment.id}`}
      status={
        <Badge variant={stateBadgeVariant(shipment.state)}>
          {(t as unknown as (k: string) => string)(`shipmentState_${shipment.state}`)}
        </Badge>
      }
      tabs={[
        {
          id: "overview",
          label: t("overviewTab"),
          content: (
            <div className="flex flex-col gap-4">
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">{t("fieldType")}:</span>
                <Badge variant="secondary">
                  {(t as unknown as (k: string) => string)(`shipmentType_${shipment.type}`)}
                </Badge>
                {shipment.origin ? (
                  <>
                    <span className="text-sm text-muted-foreground">{t("colOrigin")}:</span>
                    <span className="text-sm">{shipment.origin}</span>
                  </>
                ) : null}
              </div>
              <WorkflowSteps
                steps={shipmentSteps.map((step) => ({
                  label: (t as unknown as (k: string) => string)(`shipmentState_${step}`),
                }))}
                currentIndex={shipmentStateIndex(shipment.state)}
              />
              <div className="flex gap-2">
                {canValidate ? (
                  <Button size="sm" onClick={handleValidate} disabled={processing}>
                    {t("validate")}
                  </Button>
                ) : null}
                {canDone ? (
                  <Button size="sm" onClick={handleDone} disabled={processing}>
                    {t("markAsDone")}
                  </Button>
                ) : null}
              </div>
            </div>
          ),
        },
        {
          id: "movements",
          label: t("movementLines"),
          content: (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">{t("movementLines")}</CardTitle>
              </CardHeader>
              <CardContent>
                {movements.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t("shipmentLinesEmpty")}</p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b text-left text-muted-foreground">
                          <th className="pb-2 ">{t("fieldItem")}</th>
                          <th className="pb-2 ">{t("colQuantity")}</th>
                          <th className="pb-2 ">{t("colFrom")}</th>
                          <th className="pb-2 ">{t("colTo")}</th>
                          <th className="pb-2 ">{t("colStatus")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {movements.map((movement) => (
                          <tr key={movement.id} className="border-b last:border-0">
                            <td className="py-2">{movement.itemId}</td>
                            <td className="py-2 tabular-nums">{formatNumber(movement.qty)}</td>
                            <td className="py-2">{movement.srcLocationId}</td>
                            <td className="py-2">{movement.dstLocationId}</td>
                            <td className="py-2">
                              <Badge variant={moveStateBadgeVariant(movement.state)}>
                                {(t as unknown as (k: string) => string)(
                                  `moveState_${movement.state}`,
                                )}
                              </Badge>
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
      ]}
    />
  );
}
