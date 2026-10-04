"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Subscription } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  canActivate,
  canChurn,
  canClose,
  canPause,
  canResume,
  subscriptionStateTone,
} from "../../_components/subscription-utils";

export function SubscriptionDetail({
  orgId,
  subscriptionId,
}: {
  orgId: string;
  subscriptionId: string;
}) {
  const t = useTranslations("Subscriptions");
  const tCommon = useTranslations("Common");
  const subState = (state: string) => (t as unknown as (k: string) => string)(`state_${state}`);
  const actionLabel = (action: string) =>
    (t as unknown as (k: string) => string)(`action_${action}`);
  const queryClient = useQueryClient();

  const subQuery = useOrgQuery<{ subscription: Subscription }>(
    "subscription",
    subscriptionId,
    (organizationId) =>
      getSwantaraService().subscriptions.get(organizationId, Number(subscriptionId)),
  );

  const sub = subQuery.data?.subscription;

  const actionMutation = useMutation({
    mutationFn: (action: "activate" | "pause" | "resume" | "churn" | "close") =>
      getSwantaraService().subscriptions[action](Number(orgId), Number(subscriptionId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["subscription"] });
      void queryClient.invalidateQueries({ queryKey: ["subscriptions"] });
      toast.success(t("saved"));
    },
    onError: () => {
      toast.error(t("actionFailed"));
    },
  });

  if (subQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!sub) {
    return <p className="text-sm text-muted-foreground">{t("notFound")}</p>;
  }

  const tone = subscriptionStateTone(sub.state);

  function handleAction(action: "activate" | "pause" | "resume" | "churn" | "close") {
    if (actionMutation.isPending) return;
    actionMutation.mutate(action);
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canActivate(sub.state) ? (
        <Button
          size="sm"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("activate")}
        >
          {actionLabel("activate")}
        </Button>
      ) : null}
      {canPause(sub.state) ? (
        <Button
          size="sm"
          variant="outline"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("pause")}
        >
          {actionLabel("pause")}
        </Button>
      ) : null}
      {canResume(sub.state) ? (
        <Button
          size="sm"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("resume")}
        >
          {actionLabel("resume")}
        </Button>
      ) : null}
      {canChurn(sub.state) ? (
        <Button
          size="sm"
          variant="destructive"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("churn")}
        >
          {actionLabel("churn")}
        </Button>
      ) : null}
      {canClose(sub.state) ? (
        <Button
          size="sm"
          variant="outline"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("close")}
        >
          {actionLabel("close")}
        </Button>
      ) : null}
    </div>
  );

  const lines = sub.lines ?? [];

  return (
    <RecordLayout
      breadcrumbItems={[{ label: t("title"), href: "/subscriptions" }, { label: sub.name }]}
      title={sub.name}
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
          {subState(sub.state)}
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
                  <CardTitle>{"MRR"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatMoney(sub.mrr, { currency: DEFAULT_CURRENCY })}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("colStartDate")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {sub.dateStart ? formatDate(String(sub.dateStart)) : "—"}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("colNextInvoice")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {sub.nextInvoiceDate ? formatDate(String(sub.nextInvoiceDate)) : "—"}
                  </p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("infoTitle")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("customer")}</dt>
                      <dd className="text-sm">#{sub.contactId}</dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("plan")}</dt>
                      <dd className="text-sm">#{sub.planId}</dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("currency")}</dt>
                      <dd className="text-sm">{sub.currencyCode ?? "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("colEndDate")}</dt>
                      <dd className="text-sm">
                        {sub.dateEnd ? formatDate(String(sub.dateEnd)) : "—"}
                      </dd>
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
                        <th className="pb-2 text-left">{t("item")}</th>
                        <th className="pb-2 text-right">{t("qty")}</th>
                        <th className="pb-2 text-right">{t("price")}</th>
                        <th className="pb-2 text-right">{t("discount")}</th>
                        <th className="pb-2 text-right">{t("subtotal")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {lines.map((line) => {
                        const subtotal = line.qty * line.unitPrice * (1 - line.discountPct / 100);
                        return (
                          <tr key={line.id} className="border-b">
                            <td className="py-2">#{line.itemId}</td>
                            <td className="py-2 text-right tabular-nums">{line.qty}</td>
                            <td className="py-2 text-right tabular-nums">
                              {formatMoney(line.unitPrice, { currency: DEFAULT_CURRENCY })}
                            </td>
                            <td className="py-2 text-right tabular-nums">{line.discountPct}%</td>
                            <td className="py-2 text-right tabular-nums">
                              {formatMoney(subtotal, { currency: DEFAULT_CURRENCY })}
                            </td>
                          </tr>
                        );
                      })}
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
