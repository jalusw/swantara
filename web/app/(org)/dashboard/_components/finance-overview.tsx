"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useFinanceKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function FinanceOverview() {
  const t = useTranslations("Dashboard");
  const financeKpi = useFinanceKpi();
  const kpi = financeKpi.data?.kpi;

  const financeData = useMemo(
    () =>
      kpi
        ? [
            { label: t("stat_revenue"), value: kpi.revenue },
            { label: t("stat_expenses"), value: kpi.expenses },
            { label: t("stat_ebitda"), value: kpi.ebitda },
          ]
        : [
            { label: t("stat_revenue"), value: 0 },
            { label: t("stat_expenses"), value: 0 },
            { label: t("stat_ebitda"), value: 0 },
          ],
    [kpi, t],
  );

  const compositionData = useMemo(
    () =>
      kpi
        ? [
            { label: t("stat_expenses"), value: kpi.expenses },
            { label: t("stat_ebitda"), value: kpi.ebitda },
          ]
        : [
            { label: t("stat_expenses"), value: 0 },
            { label: t("stat_ebitda"), value: 0 },
          ],
    [kpi, t],
  );

  const marginData = useMemo(
    () =>
      kpi
        ? [
            { label: t("grossMargin"), value: kpi.grossMarginPct * 100 },
            { label: t("netMargin"), value: kpi.netMarginPct * 100 },
          ]
        : [
            { label: t("grossMargin"), value: 0 },
            { label: t("netMargin"), value: 0 },
          ],
    [kpi, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12 lg:col-span-8">
        <CardHeader>
          <CardTitle>{t("financeTitle")}</CardTitle>
          <CardDescription>{t("ledgerHealthDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={financeData}
            ariaLabel={t("financeTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-4 bg-muted/20">
        <CardHeader>
          <CardTitle className="text-sm">{t("compositionTitle")}</CardTitle>
          <CardDescription>{t("compositionDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={compositionData}
            ariaLabel={t("financeTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            centerLabel={t("stat_revenue")}
          />
        </CardContent>
      </Card>
      <div className="col-span-12 grid gap-4 grid-cols-12">
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={t("stat_revenue")}
        >
          {kpi ? formatMoney(kpi.revenue, { currency: DEFAULT_CURRENCY }) : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={t("stat_expenses")}
        >
          {kpi ? formatMoney(kpi.expenses, { currency: DEFAULT_CURRENCY }) : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
          label={t("stat_ebitda")}
        >
          {kpi ? formatMoney(kpi.ebitda, { currency: DEFAULT_CURRENCY }) : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={t("grossMargin")}
        >
          {kpi ? `${(kpi.grossMarginPct * 100).toFixed(1)}%` : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={t("netMargin")}
        >
          {kpi ? `${(kpi.netMarginPct * 100).toFixed(1)}%` : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={t("currentRatio")}
        >
          {kpi ? kpi.currentRatio.toFixed(2) : "—"}
        </KpiCard>
      </div>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{t("marginTitle")}</CardTitle>
          <CardDescription>{t("marginDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={marginData}
            ariaLabel={t("grossMargin")}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
          />
        </CardContent>
      </Card>
    </div>
  );
}
