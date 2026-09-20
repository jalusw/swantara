"use client";

import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Payment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  paymentDirectionLabel,
  paymentStateLabel,
  paymentStateTone,
} from "../../_components/payment-utils";

export function PaymentDetailSection({ paymentId }: { orgId: string; paymentId: string }) {
  const paymentQuery = useOrgQuery<{ payment: Payment }>(
    "payments",
    Number(paymentId),
    (organizationId) => getSwantaraService().payments.get(organizationId, Number(paymentId)),
  );

  const payment = paymentQuery.data?.payment;

  if (paymentQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading payment..."}</p>;
  }

  if (!payment) {
    return <p className="text-sm text-muted-foreground">{"Payment not found."}</p>;
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
          {paymentStateLabel(payment.state)}
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{"Payment summary"}</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs text-muted-foreground">{"Direction"}</dt>
              <dd className="text-sm">{paymentDirectionLabel(payment.type)}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{"Amount"}</dt>
              <dd className="text-sm tabular-nums">{formatMoney(payment.amount)}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{"Contact"}</dt>
              <dd className="text-sm">{payment.contactId ?? "—"}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{"Date"}</dt>
              <dd className="text-sm">{formatDate(payment.date)}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{"Journal"}</dt>
              <dd className="text-sm">#{payment.journalId}</dd>
            </div>
            <div>
              <dt className="text-xs text-muted-foreground">{"Payment type"}</dt>
              <dd className="text-sm">
                {payment.type === "inbound" ? "Customer payment" : "Supplier payment"}
              </dd>
            </div>
          </dl>
        </CardContent>
      </Card>
    </div>
  );
}
