"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Invoice } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { invoiceStateTone, paymentStateTone } from "../../_components/invoice-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function InvoiceDetailSection({ invoiceId }: { orgId: string; invoiceId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");
  const invoiceQuery = useOrgQuery<{ invoice: Invoice }>("invoices", invoiceId, (organizationId) =>
    getSwantaraService().invoices.get(organizationId, Number(invoiceId)),
  );

  const invoice = invoiceQuery.data?.invoice;

  if (invoiceQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!invoice) {
    return <p className="text-sm text-muted-foreground">{t("invoiceNotFound")}</p>;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold">{invoice.name ?? `INV-${invoice.id}`}</h1>
          <p className="text-sm text-muted-foreground">
            {`#${invoice.contactId}`} &middot;{" "}
            {invoice.invoiceDate ? formatDate(invoice.invoiceDate) : "—"}
          </p>
        </div>
        <div className="flex gap-2">
          <Badge variant="outline" className={invoiceStateTone(invoice.state)}>
            {(t as unknown as (k: string) => string)(`invoiceState_${invoice.state}`)}
          </Badge>
          <Badge variant="outline" className={paymentStateTone(invoice.paymentState)}>
            {(t as unknown as (k: string) => string)(`paymentState_${invoice.paymentState}`)}
          </Badge>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("summary")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{t("colStatus")}</span>
              <span className="text-sm">
                {(t as unknown as (k: string) => string)(`invoiceState_${invoice.state}`)}
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{t("colPaymentState")}</span>
              <span className="text-sm">
                {(t as unknown as (k: string) => string)(`paymentState_${invoice.paymentState}`)}
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{t("colInvoiceDate")}</span>
              <span className="text-sm">
                {invoice.invoiceDate ? formatDate(invoice.invoiceDate) : "—"}
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{t("colDueDate")}</span>
              <span className="text-sm">{invoice.dueDate ? formatDate(invoice.dueDate) : "—"}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("amounts")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 ">{t("colDescription")}</th>
                  <th className="pb-2 text-right">{t("colAmount")}</th>
                </tr>
              </thead>
              <tbody>
                <tr className="border-b">
                  <td className="py-2">{t("amountUntaxed")}</td>
                  <td className="py-2 text-right tabular-nums">
                    {formatMoney(invoice.amountUntaxed)}
                  </td>
                </tr>
                <tr className="border-b">
                  <td className="py-2">{t("amountTax")}</td>
                  <td className="py-2 text-right tabular-nums">{formatMoney(invoice.amountTax)}</td>
                </tr>
                <tr className="">
                  <td className="py-2">{t("colTotal")}</td>
                  <td className="py-2 text-right tabular-nums">
                    {formatMoney(invoice.amountTotal)}
                  </td>
                </tr>
                <tr className="text-muted-foreground">
                  <td className="py-2">{t("amountResidual")}</td>
                  <td className="py-2 text-right tabular-nums">
                    {formatMoney(invoice.amountResidual)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
