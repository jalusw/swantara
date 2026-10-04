"use client";

import { FileDown } from "lucide-react";
import { useTranslations } from "next-intl";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { useTrialBalance } from "@/lib/hooks/use-report-queries";
import { cn, exportCsv, formatNumber } from "@/lib/utils";
import { trialBalanceIsBalanced, trialBalanceTotals } from "@/lib/utils/report-utils";

import { TrialBalanceTable } from "../_components/trial-balance-table";

export default function TrialBalancePage() {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Reports");
  const { data } = useTrialBalance();
  const tb = data?.trialBalance;
  const rows = tb?.rows ?? [];
  const { totalDebit, totalCredit } = trialBalanceTotals(rows);
  const isBalanced = trialBalanceIsBalanced(rows);

  function handleExportCsv() {
    const headers = [
      t("colCode"),
      t("colAccount"),
      t("fieldType"),
      t("colOpenDr"),
      t("colOpenCr"),
      t("colPeriodDr"),
      t("colPeriodCr"),
      t("colCloseDr"),
      t("colCloseCr"),
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
      <BackLink href={"/reports"}>{t("backToReports")}</BackLink>

      <PageHeader
        title={t("trialBalanceTitle")}
        description={t("trialBalanceDescription")}
        actions={
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportCsv}
            disabled={rows.length === 0}
          >
            <FileDown className="size-4" aria-hidden />
            {t("exportCsv")}
          </Button>
        }
      />

      <section className="grid gap-3 sm:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{t("totalDebits")}</p>
            <p className="font-heading text-lg font-bold">{formatNumber(totalDebit)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{t("totalCredits")}</p>
            <p className="font-heading text-lg font-bold">{formatNumber(totalCredit)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{t("colStatus")}</p>
            <span
              className={cn(
                "inline-flex items-center rounded-full px-2 py-1 text-xs",
                isBalanced ? "bg-success/10 text-success" : "bg-destructive/10 text-destructive",
              )}
            >
              {isBalanced ? t("balanced") : t("unbalanced")}
            </span>
          </CardContent>
        </Card>
      </section>

      <TrialBalanceTable />
    </div>
  );
}
