"use client";

import { FileDown } from "lucide-react";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { useTrialBalance } from "@/lib/hooks/use-report-queries";
import { cn, exportCsv, formatNumber } from "@/lib/utils";
import { trialBalanceIsBalanced, trialBalanceTotals } from "@/lib/utils/report-utils";

import { TrialBalanceTable } from "../_components/trial-balance-table";

export default function TrialBalancePage() {
  const { data } = useTrialBalance();
  const tb = data?.trialBalance;
  const rows = tb?.rows ?? [];
  const { totalDebit, totalCredit } = trialBalanceTotals(rows);
  const isBalanced = trialBalanceIsBalanced(rows);

  function handleExportCsv() {
    const headers = [
      "Code",
      "Account",
      "Type",
      "Open. DR",
      "Open. CR",
      "Per. DR",
      "Per. CR",
      "Close. DR",
      "Close. CR",
    ];
    const csvRows = rows.map((r) => [
      r.code,
      r.name,
      r.accountType,
      r.openingDebit,
      r.openingCredit,
      r.periodDebit,
      r.periodCredit,
      r.closingDebit,
      r.closingCredit,
    ]);
    exportCsv("trial-balance", headers, csvRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{"Back to reports"}</BackLink>

      <PageHeader
        title={"Trial Balance"}
        description={"Debit and credit balances for all accounts in the current period."}
        actions={
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportCsv}
            disabled={rows.length === 0}
          >
            <FileDown className="size-4" aria-hidden />
            {"Export CSV"}
          </Button>
        }
      />

      <section className="grid gap-3 sm:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{"Total debits"}</p>
            <p className="font-heading text-lg font-bold">{formatNumber(totalDebit)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{"Total credits"}</p>
            <p className="font-heading text-lg font-bold">{formatNumber(totalCredit)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{"Status"}</p>
            <span
              className={cn(
                "inline-flex items-center rounded-full px-2 py-1 text-xs",
                isBalanced ? "bg-success/10 text-success" : "bg-destructive/10 text-destructive",
              )}
            >
              {isBalanced ? "Balanced" : "Unbalanced"}
            </span>
          </CardContent>
        </Card>
      </section>

      <TrialBalanceTable />
    </div>
  );
}
