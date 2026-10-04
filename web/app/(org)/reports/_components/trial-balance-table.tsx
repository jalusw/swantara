"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useTrialBalance } from "@/lib/hooks/use-report-queries";
import { formatMoney } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function TrialBalanceTable() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Reports");
  const { data, isLoading } = useTrialBalance();
  const tb = data?.trialBalance;
  const rows = tb?.rows ?? [];

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {t("loadingTrialBalance")}
        </CardContent>
      </Card>
    );
  }

  if (rows.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {t("trialBalanceEmpty")}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("trialBalanceTitle")}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="px-3 py-2 text-left text-muted-foreground">{t("colCode")}</th>
                <th className="px-3 py-2 text-left text-muted-foreground">{t("colAccount")}</th>
                <th className="px-3 py-2 text-left text-muted-foreground">{t("fieldType")}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{t("colOpenDr")}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{t("colOpenCr")}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{t("colPeriodDr")}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{t("colPeriodCr")}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{t("colCloseDr")}</th>
                <th className="px-3 py-2 text-right text-muted-foreground">{t("colCloseCr")}</th>
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
                  {t("colTotal")}
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
