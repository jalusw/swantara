"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { useManufacturingKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatNumber } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function ManufacturingSection() {
  const t = useTranslations("Dashboard");
  const { data } = useManufacturingKpi();
  const kpi = data?.kpi;

  const manufacturingData = useMemo(
    () =>
      kpi
        ? [
            { label: t("orders"), value: kpi.orderCount },
            { label: "OEE", value: kpi.oeePct * 100 },
            { label: t("yield"), value: kpi.yieldPct * 100 },
            { label: t("scrap"), value: kpi.scrapPct * 100 },
            { label: t("costVariance"), value: kpi.costVariancePct * 100 },
          ]
        : [
            { label: t("orders"), value: 0 },
            { label: "OEE", value: 0 },
            { label: t("yield"), value: 0 },
            { label: t("scrap"), value: 0 },
            { label: t("costVariance"), value: 0 },
          ],
    [kpi, t],
  );

  const yieldScrapData = useMemo(
    () =>
      kpi
        ? [
            { label: t("yield"), value: kpi.yieldPct * 100 },
            { label: t("scrap"), value: kpi.scrapPct * 100 },
          ]
        : [
            { label: t("yield"), value: 0 },
            { label: t("scrap"), value: 0 },
          ],
    [kpi, t],
  );

  const oeeVarianceData = useMemo(
    () =>
      kpi
        ? [
            { label: "OEE", value: kpi.oeePct * 100 },
            { label: t("costVariance"), value: kpi.costVariancePct * 100 },
          ]
        : [
            { label: "OEE", value: 0 },
            { label: t("costVariance"), value: 0 },
          ],
    [kpi, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{t("manufacturingTitle")}</CardTitle>
          <CardDescription>{t("manufacturingDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={t("orders")}
          >
            {kpi ? String(kpi.orderCount) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-2 rounded-xl border border-border bg-card p-4"
            label={"OEE"}
          >
            {kpi ? `${(kpi.oeePct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-2 rounded-xl border border-border bg-card p-4"
            label={t("yield")}
          >
            {kpi ? `${(kpi.yieldPct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-2 rounded-xl border border-border bg-card p-4"
            label={t("scrap")}
          >
            {kpi ? `${(kpi.scrapPct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-2 rounded-xl border border-border bg-card p-4"
            label={t("costVariance")}
          >
            {kpi ? `${(kpi.costVariancePct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{t("outputMetrics")}</CardTitle>
          <CardDescription>{t("outputMetricsDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={manufacturingData}
            ariaLabel={t("manufacturingTitle")}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{t("yieldVsScrap")}</CardTitle>
          <CardDescription>{t("qualityBalance")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={yieldScrapData}
            ariaLabel={t("yield")}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={t("yield")}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{t("oeeVsVariance")}</CardTitle>
          <CardDescription>{t("efficiencyVsCost")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={oeeVarianceData}
            ariaLabel={"OEE"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"OEE"}
          />
        </CardContent>
      </Card>
    </div>
  );
}
