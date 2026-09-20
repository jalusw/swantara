"use client";

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
  posOrderStateLabel,
  posOrderStateTone,
  posSessionStateConfig,
  sessionOrderTotal,
  sessionPaymentSummary,
} from "../../../_components/pos-utils";

const sessionSteps = ["opened", "closing", "closed"] as const;

function sessionStateIndex(state: PosSession["state"]): number {
  const idx = sessionSteps.indexOf(state as (typeof sessionSteps)[number]);
  return idx >= 0 ? idx : 0;
}

export function PosSessionDetail({ orgId, sessionId }: { orgId: string; sessionId: string }) {
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
    return <p className="text-sm text-muted-foreground">{"Session not found."}</p>;
  }

  const allPayments = orders.flatMap((o) => o.payments ?? []);
  const paymentSummary = sessionPaymentSummary(allPayments);
  const totalSales = sessionOrderTotal(orders);

  function handleClosing() {
    void getSwantaraService()
      .posSessions.closing(Number(orgId), session!.id)
      .then(() => {
        toast.success("Session moved to closing.");
        void sessionQuery.refetch();
      })
      .catch(() => toast.error("Something went wrong. Please try again."));
  }

  function handleClose(closingBalance: number) {
    void getSwantaraService()
      .posSessions.close(Number(orgId), session!.id, { closingBalance })
      .then(() => {
        toast.success("Session closed.");
        setClosingOpen(false);
        void sessionQuery.refetch();
      })
      .catch(() => toast.error("Something went wrong. Please try again."));
  }

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: "POS Sessions", href: "/pos/sessions" },
          { label: `Session-${session.id}` },
        ]}
        title={`Session-${session.id}`}
        status={
          <StateBadge
            value={session.state}
            statuses={{
              opened: posSessionStateConfig("opened"),
              closing: posSessionStateConfig("closing"),
              closed: posSessionStateConfig("closed"),
            }}
          />
        }
        actions={
          <div className="flex gap-2">
            {canPlaceOrder(session.state) ? (
              <Button size="sm" asChild>
                <a href={`/pos/register/${session.id}`}>{"Open register"}</a>
              </Button>
            ) : null}
            {canStartClosing(session.state) ? (
              <Button size="sm" variant="outline" onClick={handleClosing}>
                {"Start closing"}
              </Button>
            ) : null}
            {canCloseSession(session.state) ? (
              <Button size="sm" onClick={() => setClosingOpen(true)}>
                {"Close session"}
              </Button>
            ) : null}
          </div>
        }
        tabs={[
          {
            id: "overview",
            label: "Overview",
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps
                  steps={sessionSteps.map((s) => ({
                    label: String(s),
                  }))}
                  currentIndex={sessionStateIndex(session.state)}
                />
                <div className="grid gap-4 sm:grid-cols-3">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{"Opening balance"}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-lg font-bold tabular-nums">
                        {formatMoney(session.openingBalance)}
                      </p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{"Closing balance"}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-lg font-bold tabular-nums">
                        {session.closingBalance != null ? formatMoney(session.closingBalance) : "—"}
                      </p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{"Total sales"}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-lg font-bold tabular-nums">{formatMoney(totalSales)}</p>
                    </CardContent>
                  </Card>
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{"Session info"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Config"}</span>
                        <span>#{session.configId}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Cashier"}</span>
                        <span>#{session.cashierId}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Opened"}</span>
                        <span>{session.openedAt ? formatDate(session.openedAt) : "—"}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Closed"}</span>
                        <span>{session.closedAt ? formatDate(session.closedAt) : "—"}</span>
                      </div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm">{"Payment summary"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      {paymentSummary.length === 0 ? (
                        <p className="text-muted-foreground">{"No payments"}</p>
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
            label: "Orders",
            content: (
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{"Orders"}</CardTitle>
                </CardHeader>
                <CardContent>
                  {orders.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{"No orders in this session."}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-left text-muted-foreground">
                            <th className="pb-2 ">{"Order"}</th>
                            <th className="pb-2 text-right">{"Total"}</th>
                            <th className="pb-2 font-right text-right">{"Tax"}</th>
                            <th className="pb-2 ">{"State"}</th>
                            <th className="pb-2 ">{"Time"}</th>
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
                                  {posOrderStateLabel(order.state)}
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
  const [closingBalance, setClosingBalance] = useState(String(expectedBalance));
  const diff = Number(closingBalance) - expectedBalance;
  const isBalanced = Math.abs(diff) < 0.01;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Close session"}</DialogTitle>
          <DialogDescription>{"Close Description"}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Expected balance"}</span>
            <p className="text-lg font-bold tabular-nums">{formatMoney(expectedBalance)}</p>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Closing balance"}</span>
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
              {"Difference"}: {formatMoney(diff)}
            </p>
          ) : (
            <p className="text-sm text-success">{"Balanced"}</p>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={() => onSubmit(Number(closingBalance))}>{"Close session"}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
