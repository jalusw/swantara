"use client";

import { FileDown } from "lucide-react";
import { useTranslations } from "next-intl";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useBalanceSheet } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { BalanceSheetSections } from "../_components/balance-sheet-sections";

export default function BalanceSheetPage() {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Reports");
  const { data } = useBalanceSheet();
  const bs = data?.balanceSheet;

  function handleExportCsv() {
    if (!bs) return;
    const headers = [t("assets"), t("colCode"), t("colAccount"), t("fieldType"), t("colBalance")];
    const allRows = [
      ...bs.assets.map((r) => [t("assets"), r.code, r.name, r.accountType, r.balance]),
      ...bs.liabilities.map((r) => [t("liabilities"), r.code, r.name, r.accountType, r.balance]),
      ...bs.equity.map((r) => [t("equity"), r.code, r.name, r.accountType, r.balance]),
    ];
    exportCsv("balance-sheet", headers, allRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{t("backToReports")}</BackLink>

      <PageHeader
        title={t("balanceSheetTitle")}
        description={t("balanceSheetDescription")}
        actions={
          <Button variant="outline" size="sm" onClick={handleExportCsv} disabled={!bs}>
            <FileDown className="size-4" aria-hidden />
            {t("exportCsv")}
          </Button>
        }
      />

      {bs ? (
        <section className="grid gap-3 sm:grid-cols-3">
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{t("totalAssets")}</p>
              <p className="font-heading text-lg font-bold">
                {formatMoney(bs.totalAssets, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{t("totalLiabilities")}</p>
              <p className="font-heading text-lg font-bold">
                {formatMoney(bs.totalLiabilities, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{t("totalEquity")}</p>
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
