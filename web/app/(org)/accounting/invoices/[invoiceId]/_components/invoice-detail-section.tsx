"use client";

import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Invoice } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import {
  invoiceStateLabel,
  invoiceStateTone,
  paymentStateLabel,
  paymentStateTone,
} from "../../_components/invoice-utils";

export function InvoiceDetailSection({ invoiceId }: { orgId: string; invoiceId: string }) {
  const invoiceQuery = useOrgQuery<{ invoice: Invoice }>("invoices", invoiceId, (organizationId) =>
    getSwantaraService().invoices.get(organizationId, Number(invoiceId)),
  );

  const invoice = invoiceQuery.data?.invoice;

  if (invoiceQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading invoice..."}</p>;
  }

  if (!invoice) {
    return <p className="text-sm text-muted-foreground">{"Invoice not found."}</p>;
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
            {invoiceStateLabel(invoice.state)}
          </Badge>
          <Badge variant="outline" className={paymentStateTone(invoice.paymentState)}>
            {paymentStateLabel(invoice.paymentState)}
          </Badge>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{"Summary"}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{"State"}</span>
              <span className="text-sm">{invoiceStateLabel(invoice.state)}</span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{"Payment State"}</span>
              <span className="text-sm">{paymentStateLabel(invoice.paymentState)}</span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{"Invoice Date"}</span>
              <span className="text-sm">
                {invoice.invoiceDate ? formatDate(invoice.invoiceDate) : "—"}
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">{"Due date"}</span>
              <span className="text-sm">{invoice.dueDate ? formatDate(invoice.dueDate) : "—"}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{"Amounts"}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 ">{"Current balances by account group."}</th>
                  <th className="pb-2 text-right">{"Amount"}</th>
                </tr>
              </thead>
              <tbody>
                <tr className="border-b">
                  <td className="py-2">{"Untaxed"}</td>
                  <td className="py-2 text-right tabular-nums">
                    {formatMoney(invoice.amountUntaxed)}
                  </td>
                </tr>
                <tr className="border-b">
                  <td className="py-2">{"Tax"}</td>
                  <td className="py-2 text-right tabular-nums">{formatMoney(invoice.amountTax)}</td>
                </tr>
                <tr className="">
                  <td className="py-2">{"Total"}</td>
                  <td className="py-2 text-right tabular-nums">
                    {formatMoney(invoice.amountTotal)}
                  </td>
                </tr>
                <tr className="text-muted-foreground">
                  <td className="py-2">{"Residual"}</td>
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
