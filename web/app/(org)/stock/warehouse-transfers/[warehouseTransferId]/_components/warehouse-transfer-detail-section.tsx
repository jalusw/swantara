"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import { RecordLayout } from "@/components/record-layout";
import { WorkflowSteps } from "@/components/workflow-steps";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { WarehouseTransfer } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const transferSteps = ["draft", "sent", "in_transit", "received"] as const;

function transferStateIndex(state: WarehouseTransfer["state"]) {
  const idx = transferSteps.indexOf(state as (typeof transferSteps)[number]);
  return idx >= 0 ? idx : 0;
}

function stateBadgeVariant(state: WarehouseTransfer["state"]) {
  switch (state) {
    case "received":
      return "default" as const;
    case "cancelled":
      return "outline" as const;
    case "draft":
      return "secondary" as const;
    default:
      return "secondary" as const;
  }
}

export function WarehouseTransferDetail({
  orgId,
  warehouseTransferId,
}: {
  orgId: string;
  warehouseTransferId: string;
}) {
  const [processing, setProcessing] = useState(false);

  const warehouseTransferQuery = useOrgQuery<{ warehouseTransfer: WarehouseTransfer }>(
    "transferOrder",
    warehouseTransferId,
    (organizationId) =>
      getSwantaraService().inventory.transferOrder(organizationId, Number(warehouseTransferId)),
  );

  const warehouseTransfer = warehouseTransferQuery.data?.warehouseTransfer;

  if (warehouseTransferQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!warehouseTransfer) {
    return <p className="text-sm text-muted-foreground">{"Transfer not found."}</p>;
  }

  function handleSend() {
    if (!warehouseTransfer) return;
    setProcessing(true);
    void getSwantaraService()
      .inventory.sendWarehouseTransfer(Number(orgId), warehouseTransfer.id, {
        journalId: 1,
        transitAccountId: 1,
        date: new Date().toISOString(),
      })
      .then(() => void warehouseTransferQuery.refetch())
      .catch(() => toast.error("Could not disable the organization."))
      .finally(() => setProcessing(false));
  }

  function handleReceive() {
    if (!warehouseTransfer) return;
    setProcessing(true);
    void getSwantaraService()
      .inventory.receiveWarehouseTransfer(Number(orgId), warehouseTransfer.id, {
        journalId: 1,
        transitAccountId: 1,
        date: new Date().toISOString(),
      })
      .then(() => void warehouseTransferQuery.refetch())
      .catch(() => toast.error("Could not disable the organization."))
      .finally(() => setProcessing(false));
  }

  const canSend = warehouseTransfer.state === "draft";
  const canReceive = warehouseTransfer.state === "in_transit";

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "All transfers", href: "/stock/warehouse-transfers" },
        { label: warehouseTransfer.name ?? `TO-${warehouseTransfer.id}` },
      ]}
      title={warehouseTransfer.name ?? `TO-${warehouseTransfer.id}`}
      status={
        <Badge variant={stateBadgeVariant(warehouseTransfer.state)}>
          {warehouseTransfer.state}
        </Badge>
      }
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="flex flex-col gap-4">
              <WorkflowSteps
                steps={transferSteps.map((step) => ({
                  label: String(step),
                }))}
                currentIndex={transferStateIndex(warehouseTransfer.state)}
              />
              <div className="flex gap-2">
                {canSend ? (
                  <Button size="sm" onClick={handleSend} disabled={processing}>
                    {"Send"}
                  </Button>
                ) : null}
                {canReceive ? (
                  <Button size="sm" onClick={handleReceive} disabled={processing}>
                    {"Receive"}
                  </Button>
                ) : null}
              </div>
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">{"Transfer legs"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-4 text-sm">
                    <div className="flex flex-col items-center gap-1">
                      <Badge
                        variant={warehouseTransfer.state === "draft" ? "default" : "secondary"}
                      >
                        {"Out"}
                      </Badge>
                      {warehouseTransfer.outShipmentId ? (
                        <span className="text-muted-foreground">
                          PK-{warehouseTransfer.outShipmentId}
                        </span>
                      ) : null}
                    </div>
                    <span className="text-muted-foreground">→</span>
                    <div className="flex flex-col items-center gap-1">
                      <Badge
                        variant={warehouseTransfer.state === "in_transit" ? "default" : "secondary"}
                      >
                        {"Transit"}
                      </Badge>
                    </div>
                    <span className="text-muted-foreground">→</span>
                    <div className="flex flex-col items-center gap-1">
                      <Badge
                        variant={warehouseTransfer.state === "received" ? "default" : "secondary"}
                      >
                        {"In"}
                      </Badge>
                      {warehouseTransfer.inShipmentId ? (
                        <span className="text-muted-foreground">
                          PK-{warehouseTransfer.inShipmentId}
                        </span>
                      ) : null}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          ),
        },
      ]}
    />
  );
}
