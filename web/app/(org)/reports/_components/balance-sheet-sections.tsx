"use client";

import Link from "next/link";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useBalanceSheet } from "@/lib/hooks/use-report-queries";
import type { BalanceSheetAccount } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";
import type { ReportSection } from "@/lib/utils/report-utils";
import { bsGroupBySection } from "@/lib/utils/report-utils";

const sectionLabels: Record<ReportSection, string> = {
  assets: "Assets",
  liabilities: "Liabilities",
  equity: "Equity",
  revenue: "Revenue",
  cogs: "COGS",
  expenses: "Expenses",
};

export function BalanceSheetSections() {
  const { data, isLoading } = useBalanceSheet();
  const bs = data?.balanceSheet;
  const sections = bs
    ? bsGroupBySection(bs.assets, bs.liabilities, bs.equity)
    : new Map<ReportSection, BalanceSheetAccount[]>();

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {"Loading balance sheet…"}
        </CardContent>
      </Card>
    );
  }

  if (!bs) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          {"No balance sheet data available."}
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
              <CardTitle>{sectionLabels[section]}</CardTitle>
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
                    <th className="px-3 py-2 text-left text-muted-foreground">{"Code"}</th>
                    <th className="px-3 py-2 text-left text-muted-foreground">{"Account"}</th>
                    <th className="px-3 py-2 text-left text-muted-foreground">{"Type"}</th>
                    <th className="px-3 py-2 text-right text-muted-foreground">{"Balance"}</th>
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
