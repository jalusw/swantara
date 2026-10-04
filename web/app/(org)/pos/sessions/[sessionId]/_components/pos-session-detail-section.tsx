"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { RecordLayout } from "@/components/record-layout";
import { StateBadge } from "@/components/state-badge";
import { WorkflowSteps } from "@/components/workflow-steps";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { PosOrder, PosSession } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canCloseSession,
  canPlaceOrder,
  canStartClosing,
  posOrderStateTone,
  posSessionStateTone,
  sessionOrderTotal,
  sessionPaymentSummary,
} from "../../../_components/pos-utils";

const sessionSteps = ["opened", "closing", "closed"] as const;

function sessionStateIndex(state: PosSession["state"]): number {
  const idx = sessionSteps.indexOf(state as (typeof sessionSteps)[number]);
  return idx >= 0 ? idx : 0;
}

export function PosSessionDetail({ orgId, sessionId }: { orgId: string; sessionId: string }) {
  const t = useTranslations("Pos");
  const sessionState = (state: string) =>
    (t as unknown as (k: string) => string)(`sessionState_${state}`);
  const orderState = (state: string) =>
    (t as unknown as (k: string) => string)(`orderState_${state}`);
  const [closingOpen, setClosingOpen] = useState(false);

  const sessionQuery = useOrgQuery<{ session: PosSession }>(
    "posSession",
    sessionId,
    (organizationId) => getSwantaraService().posSessions.get(organizationId, Number(sessionId)),
  );

  const ordersQuery = useOrgQuery<{
    session: PosSession & { orders?: PosOrder[] };
  }>("posSession", sessionId, (organizationId) =>
    getSwantaraService().posSessions.get(organizationId, Number(sessionId)),
  );

  const session = sessionQuery.data?.session ?? null;
  const orders = ordersQuery.data?.session?.orders ?? [];

  if (sessionQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!session) {
    return <p className="text-sm text-muted-foreground">{t("sessionNotFound")}</p>;
  }

  const allPayments = orders.flatMap((o) => o.payments ?? []);
  const paymentSummary = sessionPaymentSummary(allPayments);
  const totalSales = sessionOrderTotal(orders);

  function handleClosing() {
    void getSwantaraService()
      .posSessions.closing(Number(orgId), session!.id)
      .then(() => {
        toast.success(t("movedToClosing"));
        void sessionQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleClose(closingBalance: number) {
    return getSwantaraService()
      .posSessions.close(Number(orgId), session!.id, { closingBalance })
      .then(() => {
        toast.success(t("sessionClosed"));
        setClosingOpen(false);
        void sessionQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: t("sessionsTitle"), href: "/pos/sessions" },
          { label: t("sessionFallback", { id: session.id }) },
        ]}
        title={t("sessionFallback", { id: session.id })}
        status={
          <StateBadge
            value={session.state}
            statuses={{
              opened: { label: sessionState("opened"), tone: posSessionStateTone("opened") },
              closing: { label: sessionState("closing"), tone: posSessionStateTone("closing") },
              closed: { label: sessionState("closed"), tone: posSessionStateTone("closed") },
            }}
          />
        }
        actions={
          <div className="flex gap-2">
            {canPlaceOrder(session.state) ? (
              <Button size="sm" asChild>
                <a href={`/pos/register/${session.id}`}>{t("openRegister")}</a>
              </Button>
            ) : null}
            {canStartClosing(session.state) ? (
              <Button size="sm" variant="outline" onClick={handleClosing}>
                {t("startClosing")}
              </Button>
            ) : null}
            {canCloseSession(session.state) ? (
              <Button size="sm" onClick={() => setClosingOpen(true)}>
                {t("closeSession")}
              </Button>
            ) : null}
          </div>
        }
        tabs={[
          {
            id: "overview",
            label: t("tabOverview"),
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps
                  steps={sessionSteps.map((s) => ({
                    label: sessionState(s),
                  }))}
                  currentIndex={sessionStateIndex(session.state)}
                />
                <div className="grid gap-4 sm:grid-cols-3">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{t("colOpeningBalance")}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-lg font-bold tabular-nums">
                        {formatMoney(session.openingBalance)}
                      </p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{t("colClosingBalance")}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-lg font-bold tabular-nums">
                        {session.closingBalance != null ? formatMoney(session.closingBalance) : "—"}
                      </p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{t("totalSales")}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-lg font-bold tabular-nums">{formatMoney(totalSales)}</p>
                    </CardContent>
                  </Card>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{t("sessionInfo")}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("config")}</span>
                        <span>#{session.configId}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("cashier")}</span>
                        <span>#{session.cashierId}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("colOpenedAt")}</span>
                        <span>{session.openedAt ? formatDate(session.openedAt) : "—"}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("closedAt")}</span>
                        <span>{session.closedAt ? formatDate(session.closedAt) : "—"}</span>
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{t("paymentSummary")}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      {paymentSummary.length === 0 ? (
                        <p className="text-muted-foreground">{t("noPayments")}</p>
                      ) : (
                        paymentSummary.map((ps) => (
                          <div key={ps.method} className="flex justify-between">
                            <span className="text-muted-foreground">{ps.method}</span>
                            <span className="tabular-nums ">{formatMoney(ps.total)}</span>
                          </div>
                        ))
                      )}
                    </CardContent>
                  </Card>
                </div>
              </div>
            ),
          },
          {
            id: "orders",
            label: t("tabOrders"),
            content: (
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("tabOrders")}</CardTitle>
                </CardHeader>
                <CardContent>
                  {orders.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t("noSessionOrders")}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-left text-muted-foreground">
                            <th className="pb-2 ">{t("colOrder")}</th>
                            <th className="pb-2 text-right">{t("colTotal")}</th>
                            <th className="pb-2 font-right text-right">{t("colTax")}</th>
                            <th className="pb-2 ">{t("colStatus")}</th>
                            <th className="pb-2 ">{t("colTime")}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {orders.map((order) => (
                            <tr key={order.id} className="border-b last:border-0">
                              <td className="py-2">
                                <a
                                  href={`/pos/orders/${order.id}`}
                                  className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                >
                                  {order.name ?? `POS-${order.id}`}
                                </a>
                              </td>
                              <td className="py-2 text-right tabular-nums">
                                {formatMoney(order.amountTotal)}
                              </td>
                              <td className="py-2 text-right tabular-nums">
                                {formatMoney(order.amountTax)}
                              </td>
                              <td className="py-2">
                                <Badge variant="outline" className={posOrderStateTone(order.state)}>
                                  {orderState(order.state)}
                                </Badge>
                              </td>
                              <td className="py-2 text-muted-foreground">
                                {order.orderTime ? formatDate(order.orderTime) : "—"}
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
      {closingOpen ? (
        <ClosingDialog
          open={closingOpen}
          onOpenChange={setClosingOpen}
          expectedBalance={totalSales + session.openingBalance}
          onSubmit={handleClose}
        />
      ) : null}
    </>
  );
}

function ClosingDialog({
  open,
  onOpenChange,
  expectedBalance,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  expectedBalance: number;
  onSubmit: (closingBalance: number) => void;
}) {
  const t = useTranslations("Pos");
  const tCommon = useTranslations("Common");
  const [closingBalance, setClosingBalance] = useState(String(expectedBalance));
  const diff = Number(closingBalance) - expectedBalance;
  const isBalanced = Math.abs(diff) < 0.01;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("closeSession")}</DialogTitle>
          <DialogDescription>{t("closeDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("expectedBalance")}</span>
            <p className="text-lg font-bold tabular-nums">{formatMoney(expectedBalance)}</p>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("colClosingBalance")}</span>
            <Input
              type="number"
              step="any"
              min="0"
              value={closingBalance}
              onChange={(e) => setClosingBalance(e.target.value)}
            />
          </div>
          {!isBalanced ? (
            <p className="text-sm text-destructive">
              {t("difference")}: {formatMoney(diff)}
            </p>
          ) : (
            <p className="text-sm text-success">{t("balanced")}</p>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={() => onSubmit(Number(closingBalance))}>{t("closeSession")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
