"use client";

import Link from "next/link";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useTrialBalance } from "@/lib/hooks/use-report-queries";
import { formatMoney } from "@/lib/utils";

export function TrialBalanceTable() {
  const { data, isLoading } = useTrialBalance();
  const tb = data?.trialBalance;
  const rows = tb?.rows ?? [];

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {"Loading trial balance…"}
        </CardContent>
      </Card>
    );
  }

  if (rows.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {"No trial balance data available."}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Trial balance"}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-3 py-2 text-left text-muted-foreground">{"Code"}</th>
                <th className="px-3 py-2 text-left text-muted-foreground">{"Account"}</th>
                <th className="px-3 py-2 text-left text-muted-foreground">{"Type"}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{"Open. DR"}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{"Open. CR"}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{"Per. DR"}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{"Per. CR"}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{"Close. DR"}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{"Close. CR"}</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.accountId} className="border-b border-border/50">
                  <td className="px-3 py-2 font-mono text-xs">
                    <Link
                      href={`/accounts/${row.accountId}`}
                      className="text-primary hover:underline"
                    >
                      {row.code}
                    </Link>
                  </td>
                  <td className="px-3 py-2">{row.name}</td>
                  <td className="px-3 py-2">
                    <Badge variant="outline">{row.accountType}</Badge>
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums">
                    {row.openingDebit
                      ? formatMoney(row.openingDebit, { currency: DEFAULT_CURRENCY })
                      : "—"}
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums">
                    {row.openingCredit
                      ? formatMoney(row.openingCredit, { currency: DEFAULT_CURRENCY })
                      : "—"}
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums">
                    {row.periodDebit
                      ? formatMoney(row.periodDebit, { currency: DEFAULT_CURRENCY })
                      : "—"}
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums">
                    {row.periodCredit
                      ? formatMoney(row.periodCredit, { currency: DEFAULT_CURRENCY })
                      : "—"}
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums ">
                    {formatMoney(row.closingDebit, { currency: DEFAULT_CURRENCY })}
                  </td>
                  <td className="px-3 py-2 text-right tabular-nums ">
                    {formatMoney(row.closingCredit, { currency: DEFAULT_CURRENCY })}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr className="border-t-2 border-border ">
                <td className="px-3 py-2" colSpan={3}>
                  {"Totals"}
                </td>
                <td className="px-3 py-2 text-right tabular-nums">
                  {formatMoney(
                    rows.reduce((s, r) => s + r.openingDebit, 0),
                    { currency: DEFAULT_CURRENCY },
                  )}
                </td>
                <td className="px-3 py-2 text-right tabular-nums">
                  {formatMoney(
                    rows.reduce((s, r) => s + r.openingCredit, 0),
                    { currency: DEFAULT_CURRENCY },
                  )}
                </td>
                <td className="px-3 py-2 text-right tabular-nums">
                  {formatMoney(
                    rows.reduce((s, r) => s + r.periodDebit, 0),
                    { currency: DEFAULT_CURRENCY },
                  )}
                </td>
                <td className="px-3 py-2 text-right tabular-nums">
                  {formatMoney(
                    rows.reduce((s, r) => s + r.periodCredit, 0),
                    { currency: DEFAULT_CURRENCY },
                  )}
                </td>
                <td className="px-3 py-2 text-right tabular-nums">
                  {formatMoney(
                    rows.reduce((s, r) => s + r.closingDebit, 0),
                    { currency: DEFAULT_CURRENCY },
                  )}
                </td>
                <td className="px-3 py-2 text-right tabular-nums">
                  {formatMoney(
                    rows.reduce((s, r) => s + r.closingCredit, 0),
                    { currency: DEFAULT_CURRENCY },
                  )}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}
