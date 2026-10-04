"use client";

import { FileDown } from "lucide-react";
import { useTranslations } from "next-intl";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useCashFlowReport } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { cn } from "@/lib/utils/style";

export default function CashFlowPage() {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Reports");
  const { data, isLoading } = useCashFlowReport();
  const cf = data?.cashFlow;

  function handleExportCsv() {
    if (!cf) return;
    const headers = [
      t("operating"),
      t("investing"),
      t("financing"),
      t("netChange"),
      t("openingCash"),
      t("closingCash"),
    ];
    const csvRows = [
      [cf.operating, cf.investing, cf.financing, cf.netChange, cf.openingCash, cf.closingCash],
    ];
    exportCsv("cash-flow", headers, csvRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{t("backToReports")}</BackLink>

      <PageHeader
        title={t("cashFlowTitle")}
        description={t("cashFlowDescription")}
        actions={
          <Button variant="outline" size="sm" onClick={handleExportCsv} disabled={!cf}>
            <FileDown className="size-4" aria-hidden />
            {t("exportCsv")}
          </Button>
        }
      />

      {isLoading ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {t("loadingCashFlow")}
          </CardContent>
        </Card>
      ) : !cf ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {t("cashFlowEmpty")}
          </CardContent>
        </Card>
      ) : (
        <>
          <section className="grid gap-3 sm:grid-cols-3">
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{t("operating")}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.operating >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.operating, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{t("investing")}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.investing >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.investing, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{t("financing")}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.financing >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.financing, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
          </section>

          <section className="grid gap-3 sm:grid-cols-3">
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{t("netChange")}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.netChange >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.netChange, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{t("openingCash")}</p>
                <p className="font-heading text-lg font-bold">
                  {formatMoney(cf.openingCash, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{t("closingCash")}</p>
                <p className="font-heading text-lg font-bold">
                  {formatMoney(cf.closingCash, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
          </section>
        </>
      )}
    </div>
  );
}
