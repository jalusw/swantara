"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useBalanceSheet } from "@/lib/hooks/use-report-queries";
import type { BalanceSheetAccount } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";
import type { ReportSection } from "@/lib/utils/report-utils";
import { bsGroupBySection } from "@/lib/utils/report-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function BalanceSheetSections() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Reports");
  const { data, isLoading } = useBalanceSheet();
  const bs = data?.balanceSheet;
  const sections = bs
    ? bsGroupBySection(bs.assets, bs.liabilities, bs.equity)
    : new Map<ReportSection, BalanceSheetAccount[]>();

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {t("loadingBalanceSheet")}
        </CardContent>
      </Card>
    );
  }

  if (!bs) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {t("balanceSheetEmpty")}
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {Array.from(sections.entries()).map(([section, sectionRows]) => (
        <Card key={section}>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>{(t as unknown as (k: string) => string)(`section_${section}`)}</CardTitle>
              <Badge variant="outline">
                {formatMoney(sectionRows.reduce((s, r) => s + r.balance, 0))}
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
                    <th className="px-3 py-2 text-left text-muted-foreground">{t("fieldType")}</th>
                    <th className="px-3 py-2 text-right text-muted-foreground">
                      {t("colBalance")}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {sectionRows.map((row) => (
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
                      <td className="px-3 py-2 text-right tabular-nums ">
                        {formatMoney(row.balance, { currency: DEFAULT_CURRENCY })}
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
