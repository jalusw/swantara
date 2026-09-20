"use client";

import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { InboundCost, InboundCostAdjustment, InboundCostLine } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { canCancel, canPost, inboundCostStateTone } from "../../_components/inbound-cost-utils";

export function InboundCostDetail({
  orgId,
  inboundCostId,
}: {
  orgId: string;
  inboundCostId: string;
}) {
  const costQuery = useOrgListQuery<InboundCost, Record<string, never>>(
    "inboundCost",
    (organizationId) =>
      getSwantaraService().inventory.inboundCost(organizationId, Number(inboundCostId)),
  );

  const linesQuery = useOrgListQuery<{ lines: InboundCostLine[] }, Record<string, never>>(
    "inboundCostLines",
    (organizationId) =>
      getSwantaraService().inventory.inboundCostLines(organizationId, Number(inboundCostId)),
  );

  const adjustmentsQuery = useOrgListQuery<
    { adjustments: InboundCostAdjustment[] },
    Record<string, never>
  >("inboundCostAdjustments", (organizationId) =>
    getSwantaraService().inventory.inboundCostAdjustments(organizationId, Number(inboundCostId)),
  );

  const cost = costQuery.data ?? null;
  const lines = linesQuery.data?.lines ?? [];
  const adjustments = adjustmentsQuery.data?.adjustments ?? [];

  if (costQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!cost) {
    return <p className="text-sm text-muted-foreground">{"Landed cost not found."}</p>;
  }

  const tone = inboundCostStateTone(cost.state);

  function handlePost() {
    void toast.promise(
      getSwantaraService().inventory.postInboundCost(Number(orgId), Number(inboundCostId)),
      {
        loading: "Processing…",
        success: () => {
          void costQuery.refetch();
          void linesQuery.refetch();
          void adjustmentsQuery.refetch();
          return "Landed cost posted";
        },
        error: "Action failed",
      },
    );
  }

  function handleCancel() {
    void toast.promise(
      getSwantaraService().inventory.postInboundCost(Number(orgId), Number(inboundCostId)),
      {
        loading: "Processing…",
        success: () => {
          void costQuery.refetch();
          return "Landed cost cancelled";
        },
        error: "Action failed",
      },
    );
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canPost(cost.state) ? (
        <Button size="sm" onClick={handlePost}>
          {"Post"}
        </Button>
      ) : null}
      {canCancel(cost.state) ? (
        <Button size="sm" variant="destructive" onClick={handleCancel}>
          {"Cancel"}
        </Button>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Landed costs", href: "/purchases/inbound-costs" },
        { label: cost.name },
      ]}
      title={cost.name}
      status={
        <Badge
          variant="outline"
          className={
            tone === "success"
              ? "border-success text-success"
              : tone === "warning"
                ? "border-warning text-warning"
                : tone === "info"
                  ? "border-info text-info"
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : ""
          }
        >
          {humanizeKey(String(cost.state))}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Date"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{cost.date ? formatDate(String(cost.date)) : "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Move"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{cost.movementId ? `#${cost.movementId}` : "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Target shipments"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {cost.targetShipmentIds.length > 0
                      ? cost.targetShipmentIds.map((id) => `#${id}`).join(", ")
                      : "—"}
                  </p>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "lines",
          label: "Cost lines",
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{"No cost lines."}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{"Item"}</th>
                        <th className="pb-2 text-left">{"Description"}</th>
                        <th className="pb-2 text-right">{"Amount"}</th>
                        <th className="pb-2 text-right">{"Split method"}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {lines.map((line) => (
                        <tr key={line.id} className="border-b">
                          <td className="py-2">#{line.itemId}</td>
                          <td className="py-2 text-muted-foreground">{line.description ?? "—"}</td>
                          <td className="py-2 text-right tabular-nums ">
                            {formatNumber(line.amount)}
                          </td>
                          <td className="py-2 text-right text-muted-foreground">
                            {String(line.splitMethod)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ),
        },
        {
          id: "adjustments",
          label: "Adjustments",
          content: (
            <div className="space-y-4">
              {adjustments.length === 0 ? (
                <p className="text-sm text-muted-foreground">{"No adjustments."}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{"Item"}</th>
                        <th className="pb-2 text-left">{"Stock Move"}</th>
                        <th className="pb-2 text-right">{"Additional Cost"}</th>
                        <th className="pb-2 text-right">{"Valuation Layer"}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {adjustments.map((adj) => (
                        <tr key={adj.id} className="border-b">
                          <td className="py-2">#{adj.itemId}</td>
                          <td className="py-2 text-muted-foreground">#{adj.stockMovementId}</td>
                          <td className="py-2 text-right tabular-nums ">
                            {formatNumber(adj.additionalCost)}
                          </td>
                          <td className="py-2 text-right text-muted-foreground">
                            #{adj.valuationLayerId}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ),
        },
      ]}
    />
  );
}
