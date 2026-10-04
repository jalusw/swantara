"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { useProcurementKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatNumber } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function ProcurementKpiCard() {
  const t = useTranslations("Dashboard");
  const { data } = useProcurementKpi();
  const kpi = data?.kpi;

  const procurementData = useMemo(
    () =>
      kpi
        ? [
            { label: t("purchaseOrders"), value: kpi.purchaseCount },
            { label: t("avgCycleDays"), value: kpi.avgCycleDays },
            { label: t("onTime"), value: kpi.onTimeDeliveryPct * 100 },
            { label: t("priceVariance"), value: kpi.priceVariancePct * 100 },
          ]
        : [
            { label: t("purchaseOrders"), value: 0 },
            { label: t("avgCycleDays"), value: 0 },
            { label: t("onTime"), value: 0 },
            { label: t("priceVariance"), value: 0 },
          ],
    [kpi, t],
  );

  const deliveryData = useMemo(
    () =>
      kpi
        ? [
            { label: t("onTime"), value: kpi.onTimeDeliveryPct * 100 },
            { label: t("late"), value: (1 - kpi.onTimeDeliveryPct) * 100 },
          ]
        : [
            { label: t("onTime"), value: 0 },
            { label: t("late"), value: 0 },
          ],
    [kpi, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{t("procurementTitle")}</CardTitle>
          <CardDescription>{t("procurementDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4"
            label={t("purchaseOrders")}
          >
            {kpi ? String(kpi.purchaseCount) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4"
            label={t("avgCycleDays")}
          >
            {kpi ? `${kpi.avgCycleDays.toFixed(1)}d` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={t("onTime")}
          >
            {kpi ? `${(kpi.onTimeDeliveryPct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4"
            label={t("priceVariance")}
          >
            {kpi ? `${(kpi.priceVariancePct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-7">
        <CardHeader>
          <CardTitle className="text-sm">{t("procurementMetrics")}</CardTitle>
          <CardDescription>{t("procurementMetricsDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={procurementData}
            ariaLabel={t("procurementTitle")}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-5">
        <CardHeader>
          <CardTitle className="text-sm">{t("onTimeDelivery")}</CardTitle>
          <CardDescription>{t("onTimeVsLate")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={deliveryData}
            ariaLabel={t("onTime")}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={t("onTime")}
          />
        </CardContent>
      </Card>
    </div>
  );
}
