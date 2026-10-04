"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useProfitAndLoss } from "@/lib/hooks/use-report-queries";
import { formatMoney } from "@/lib/utils";
import { plGroupByType } from "@/lib/utils/report-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function ProfitAndLossRows() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Reports");
  const { data, isLoading } = useProfitAndLoss();
  const pl = data?.profitAndLoss;
  const rows = pl?.rows ?? [];
  const grouped = plGroupByType(rows);

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {t("loadingProfitAndLoss")}
        </CardContent>
      </Card>
    );
  }

  if (rows.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {t("profitAndLossEmpty")}
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {Array.from(grouped.entries()).map(([type, typeRows]) => (
        <Card key={type}>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>{type}</CardTitle>
              <Badge variant="outline">
                {formatMoney(typeRows.reduce((s, r) => s + r.amount, 0))}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="px-3 py-2 text-left text-muted-foreground">{t("colCode")}</th>
                    <th className="px-3 py-2 text-left text-muted-foreground">{t("colAccount")}</th>
                    <th className="px-3 py-2 text-right text-muted-foreground">{t("colAmount")}</th>
                  </tr>
                </thead>
                <tbody>
                  {typeRows.map((row) => (
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
                      <td className="px-3 py-2 text-right tabular-nums ">
                        {formatMoney(row.amount, { currency: DEFAULT_CURRENCY })}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
