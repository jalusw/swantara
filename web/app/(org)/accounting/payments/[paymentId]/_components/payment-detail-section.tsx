"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Payment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { paymentStateTone } from "../../_components/payment-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function PaymentDetailSection({ paymentId }: { orgId: string; paymentId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");
  const paymentQuery = useOrgQuery<{ payment: Payment }>(
    "payments",
    Number(paymentId),
    (organizationId) => getSwantaraService().payments.get(organizationId, Number(paymentId)),
  );

  const payment = paymentQuery.data?.payment;

  if (paymentQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!payment) {
    return <p className="text-sm text-muted-foreground">{t("paymentNotFound")}</p>;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold">{payment.name ?? `PAY-${payment.id}`}</h1>
          <p className="text-sm text-muted-foreground">
            {payment.contactId} &middot; {formatDate(payment.date)}
          </p>
        </div>
        <Badge variant="outline" className={paymentStateTone(payment.state)}>
          {(t as unknown as (k: string) => string)(`paymentState_${payment.state}`)}
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("paymentSummary")}</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs text-muted-foreground">{t("colDirection")}</dt>
              <dd className="text-sm">
                {(t as unknown as (k: string) => string)(`paymentDirection_${payment.type}`)}
              </dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{t("colAmount")}</dt>
              <dd className="text-sm tabular-nums">{formatMoney(payment.amount)}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{t("colContact")}</dt>
              <dd className="text-sm">{payment.contactId ?? "—"}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{t("colDate")}</dt>
              <dd className="text-sm">{formatDate(payment.date)}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{t("fieldJournal")}</dt>
              <dd className="text-sm">#{payment.journalId}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{t("colPaymentType")}</dt>
              <dd className="text-sm">
                {payment.type === "inbound" ? t("customerPayment") : t("supplierPayment")}
              </dd>
            </div>
          </dl>
        </CardContent>
      </Card>
    </div>
  );
}
