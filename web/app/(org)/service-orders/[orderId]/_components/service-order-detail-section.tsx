"use client";

import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { ServiceOrder, ServiceOrderLine } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import {
  canBill,
  canCancel,
  canComplete,
  canSchedule,
  canStart,
  serviceOrderStateTone,
} from "../../_components/service-order-utils";

export function ServiceOrderDetail({ orgId, orderId }: { orgId: string; orderId: string }) {
  const t = useTranslations("Service");
  const tCommon = useTranslations("Common");
  const orderState = (state: string) =>
    (t as unknown as (k: string) => string)(`orderState_${state}`);
  const orderType = (type: string) => (t as unknown as (k: string) => string)(`type_${type}`);
  const lineType = (type: string) => (t as unknown as (k: string) => string)(`lineType_${type}`);
  const actionLabel = (action: string) =>
    (t as unknown as (k: string) => string)(`detailAction_${action}`);
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
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!order) {
    return <p className="text-sm text-muted-foreground">{t("orderNotFound")}</p>;
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
      loading: t("processing"),
      success: () => {
        void query.refetch();
        return t("saved");
      },
      error: t("actionFailed"),
    });
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canSchedule(order.state) ? (
        <Button size="sm" onClick={() => handleAction("schedule")}>
          {actionLabel("schedule")}
        </Button>
      ) : null}
      {canStart(order.state) ? (
        <Button size="sm" onClick={() => handleAction("start")}>
          {actionLabel("start")}
        </Button>
      ) : null}
      {canComplete(order.state) ? (
        <Button size="sm" onClick={() => handleAction("complete")}>
          {actionLabel("complete")}
        </Button>
      ) : null}
      {canBill(order.state) ? (
        <Button size="sm" onClick={() => handleAction("bill")}>
          {actionLabel("bill")}
        </Button>
      ) : null}
      {canCancel(order.state) ? (
        <Button size="sm" variant="destructive" onClick={() => handleAction("cancel")}>
          {tCommon("cancel")}
        </Button>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("ordersTitle"), href: "/service-orders" },
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
          {orderState(order.state)}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: t("tabOverview"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("colType")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{orderType(order.type)}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("colPriority")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm tabular-nums">{order.priority}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("scheduledDate")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {order.scheduledDate ? formatDate(String(order.scheduledDate)) : "—"}
                  </p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("detailsTitle")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("customer")}</dt>
                      <dd className="text-sm">{order.contactId ? `#${order.contactId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("equipment")}</dt>
                      <dd className="text-sm">
                        {order.equipmentId ? `#${order.equipmentId}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("contract")}</dt>
                      <dd className="text-sm">{order.contractId ? `#${order.contractId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("technician")}</dt>
                      <dd className="text-sm">
                        {order.technicianId ? `#${order.technicianId}` : "—"}
                      </dd>
                    </div>
                    <div className="sm:col-span-2">
                      <dt className="text-sm text-muted-foreground">{t("reportedIssue")}</dt>
                      <dd className="text-sm">{order.reportedIssue || "—"}</dd>
                    </div>
                    <div className="sm:col-span-2">
                      <dt className="text-sm text-muted-foreground">{t("resolution")}</dt>
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
          label: t("tabLines"),
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{t("noLines")}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{t("colType")}</th>
                        <th className="pb-2 text-left">{t("description")}</th>
                        <th className="pb-2 text-right">{t("qty")}</th>
                        <th className="pb-2 text-right">{t("unitCost")}</th>
                        <th className="pb-2 text-right">{t("price")}</th>
                        <th className="pb-2 text-right">{t("billable")}</th>
                        <th className="pb-2 text-right">{t("warranty")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {lines.map((line) => (
                        <tr key={line.id} className="border-b">
                          <td className="py-2">
                            <Badge variant="outline">{lineType(line.type)}</Badge>
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
