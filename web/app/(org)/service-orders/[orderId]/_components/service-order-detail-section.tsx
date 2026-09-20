"use client";

import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { ServiceOrder, ServiceOrderLine } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import {
  canBill,
  canCancel,
  canComplete,
  canSchedule,
  canStart,
  serviceOrderStateTone,
} from "../../_components/service-order-utils";

export function ServiceOrderDetail({ orgId, orderId }: { orgId: string; orderId: string }) {
  const query = useOrgQuery<{ serviceOrder: ServiceOrder }>(
    "serviceOrder",
    orderId,
    (organizationId) => getSwantaraService().serviceOrders.get(organizationId, Number(orderId)),
  );

  const linesQuery = useOrgListQuery<
    { serviceOrderLines: ServiceOrderLine[] },
    Record<string, never>
  >("serviceOrderLines", (organizationId) =>
    getSwantaraService().serviceOrders.lines.list(organizationId, Number(orderId)),
  );

  const order = query.data?.serviceOrder;
  const lines = linesQuery.data?.serviceOrderLines ?? [];

  if (query.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!order) {
    return <p className="text-sm text-muted-foreground">{"Service order not found."}</p>;
  }

  const tone = serviceOrderStateTone(order.state);

  function handleAction(action: "schedule" | "start" | "complete" | "bill" | "cancel") {
    const svc = getSwantaraService().serviceOrders;
    const org = Number(orgId);
    const id = Number(orderId);
    let promise: Promise<unknown>;
    switch (action) {
      case "schedule":
        promise = svc.schedule(org, id, { scheduledDate: new Date().toISOString() });
        break;
      case "start":
        promise = svc.start(org, id);
        break;
      case "complete":
        promise = svc.complete(org, id, {
          journalId: 1,
          date: new Date().toISOString(),
          cogsAccountId: 1,
          stockCostAccountId: 1,
        });
        break;
      case "bill":
        promise = svc.bill(org, id, { journalId: 1, revenueAccountId: 1 });
        break;
      default:
        promise = svc.cancel(org, id);
        break;
    }
    void toast.promise(promise, {
      loading: "Processing…",
      success: () => {
        void query.refetch();
        return "Saved.";
      },
      error: "Action failed",
    });
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canSchedule(order.state) ? (
        <Button size="sm" onClick={() => handleAction("schedule")}>
          {"Schedule"}
        </Button>
      ) : null}
      {canStart(order.state) ? (
        <Button size="sm" onClick={() => handleAction("start")}>
          {"Start"}
        </Button>
      ) : null}
      {canComplete(order.state) ? (
        <Button size="sm" onClick={() => handleAction("complete")}>
          {"Complete"}
        </Button>
      ) : null}
      {canBill(order.state) ? (
        <Button size="sm" onClick={() => handleAction("bill")}>
          {"Bill"}
        </Button>
      ) : null}
      {canCancel(order.state) ? (
        <Button size="sm" variant="destructive" onClick={() => handleAction("cancel")}>
          {"Cancel"}
        </Button>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Service orders", href: "/service-orders" },
        { label: order.name },
      ]}
      title={order.name}
      status={
        <Badge
          variant="outline"
          className={
            tone === "success"
              ? "border-success text-success"
              : tone === "warning"
                ? "border-warning text-warning"
                : tone === "danger"
                  ? "border-destructive text-destructive"
                  : tone === "info"
                    ? "border-info text-info"
                    : ""
          }
        >
          {humanizeKey(String(order.state))}
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
                  <CardTitle>{"Type"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{humanizeKey(String(order.type))}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Priority"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm tabular-nums">{order.priority}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Scheduled date"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {order.scheduledDate ? formatDate(String(order.scheduledDate)) : "—"}
                  </p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{"Details"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Customer"}</dt>
                      <dd className="text-sm">{order.contactId ? `#${order.contactId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Equipment"}</dt>
                      <dd className="text-sm">
                        {order.equipmentId ? `#${order.equipmentId}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Contract"}</dt>
                      <dd className="text-sm">{order.contractId ? `#${order.contractId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Technician"}</dt>
                      <dd className="text-sm">
                        {order.technicianId ? `#${order.technicianId}` : "—"}
                      </dd>
                    </div>
                    <div className="sm:col-span-2">
                      <dt className="text-sm text-muted-foreground">{"Reported issue"}</dt>
                      <dd className="text-sm">{order.reportedIssue || "—"}</dd>
                    </div>
                    <div className="sm:col-span-2">
                      <dt className="text-sm text-muted-foreground">{"Resolution"}</dt>
                      <dd className="text-sm">{order.resolution || "—"}</dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "lines",
          label: "Lines",
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{"No lines."}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{"Type"}</th>
                        <th className="pb-2 text-left">{"Description"}</th>
                        <th className="pb-2 text-right">{"Qty"}</th>
                        <th className="pb-2 text-right">{"Cost"}</th>
                        <th className="pb-2 text-right">{"Price"}</th>
                        <th className="pb-2 text-right">{"Billable"}</th>
                        <th className="pb-2 text-right">{"Warranty"}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {lines.map((line) => (
                        <tr key={line.id} className="border-b">
                          <td className="py-2">
                            <Badge variant="outline">{humanizeKey(String(line.type))}</Badge>
                          </td>
                          <td className="py-2 text-muted-foreground">{line.description || "—"}</td>
                          <td className="py-2 text-right tabular-nums">{line.qty}</td>
                          <td className="py-2 text-right tabular-nums">
                            {formatNumber(line.unitCost)}
                          </td>
                          <td className="py-2 text-right tabular-nums ">
                            {formatNumber(line.unitPrice)}
                          </td>
                          <td className="py-2 text-right">{line.billable ? "✓" : "—"}</td>
                          <td className="py-2 text-right">{line.coveredByWarranty ? "✓" : "—"}</td>
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
