"use client";

import { FileDown } from "lucide-react";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useBalanceSheet } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { BalanceSheetSections } from "../_components/balance-sheet-sections";

export default function BalanceSheetPage() {
  const { data } = useBalanceSheet();
  const bs = data?.balanceSheet;

  function handleExportCsv() {
    if (!bs) return;
    const headers = ["Assets", "Code", "Account", "Type", "Balance"];
    const allRows = [
      ...bs.assets.map((r) => ["Assets", r.code, r.name, r.accountType, r.balance]),
      ...bs.liabilities.map((r) => ["Liabilities", r.code, r.name, r.accountType, r.balance]),
      ...bs.equity.map((r) => ["Equity", r.code, r.name, r.accountType, r.balance]),
    ];
    exportCsv("balance-sheet", headers, allRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{"Back to reports"}</BackLink>

      <PageHeader
        title={"Balance Sheet"}
        description={"Assets, liabilities, and equity as of today."}
        actions={
          <Button variant="outline" size="sm" onClick={handleExportCsv} disabled={!bs}>
            <FileDown className="size-4" aria-hidden />
            {"Export CSV"}
          </Button>
        }
      />

      {bs ? (
        <section className="grid gap-3 sm:grid-cols-3">
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{"Total assets"}</p>
              <p className="font-heading text-lg font-bold">
                {formatMoney(bs.totalAssets, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{"Total liabilities"}</p>
              <p className="font-heading text-lg font-bold">
                {formatMoney(bs.totalLiabilities, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{"Total equity"}</p>
              <p className="font-heading text-lg font-bold">
                {formatMoney(bs.totalEquity, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
        </section>
      ) : null}

      <BalanceSheetSections />
    </div>
  );
}
