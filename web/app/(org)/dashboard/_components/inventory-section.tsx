"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { LineChart } from "@/components/line-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useInventoryKpi, useInventoryRatioKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney, formatNumber } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function InventoryOverview() {
  const t = useTranslations("Dashboard");
  const inventoryKpi = useInventoryKpi();
  const inv = inventoryKpi.data?.kpi;

  const inventoryData = useMemo(
    () =>
      inv
        ? [
            { label: t("onHandValue"), value: inv.onHandValue },
            { label: t("onHandQty"), value: inv.onHandQuantity },
          ]
        : [
            { label: t("onHandValue"), value: 0 },
            { label: t("onHandQty"), value: 0 },
          ],
    [inv, t],
  );

  const stockComposition = useMemo(
    () =>
      inv
        ? [
            { label: t("onHandQty"), value: inv.onHandQuantity },
            { label: t("products"), value: inv.productCount },
          ]
        : [
            { label: t("onHandQty"), value: 0 },
            { label: t("products"), value: 0 },
          ],
    [inv, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{t("inventoryTitle")}</CardTitle>
          <CardDescription>{t("inventoryDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-12 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={t("onHandValue")}
          >
            {inv ? formatMoney(inv.onHandValue, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("onHandQty")}
          >
            {inv ? formatNumber(inv.onHandQuantity) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("products")}
          >
            {inv ? formatNumber(inv.productCount) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-7">
        <CardHeader>
          <CardTitle className="text-sm">{t("onHandBreakdown")}</CardTitle>
          <CardDescription>{t("valueVsQty")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={inventoryData}
            ariaLabel={t("inventoryTitle")}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-5">
        <CardHeader>
          <CardTitle className="text-sm">{t("catalogBreadth")}</CardTitle>
          <CardDescription>{t("qtyVsProducts")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={stockComposition}
            ariaLabel={t("inventoryTitle")}
            valueFormatter={(value) => formatNumber(value)}
            centerLabel={t("products")}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export function InventoryHealthSection() {
  const t = useTranslations("Dashboard");
  const ratioKpi = useInventoryRatioKpi();
  const ratio = ratioKpi.data?.kpi;

  const ratioData = useMemo(
    () =>
      ratio
        ? [
            { label: t("turnover"), value: ratio.turnover },
            { label: t("daysOnHand"), value: ratio.daysOnHand },
            { label: t("stockouts"), value: ratio.stockoutCount },
          ]
        : [
            { label: t("turnover"), value: 0 },
            { label: t("daysOnHand"), value: 0 },
            { label: t("stockouts"), value: 0 },
          ],
    [ratio, t],
  );

  const healthFlow = useMemo(
    () =>
      ratio
        ? [
            { label: t("turnover"), value: ratio.turnover },
            { label: t("daysOnHand"), value: ratio.daysOnHand },
          ]
        : [
            { label: t("turnover"), value: 0 },
            { label: t("daysOnHand"), value: 0 },
          ],
    [ratio, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{t("healthTitle")}</CardTitle>
          <CardDescription>{t("healthDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
            label={t("turnover")}
          >
            {ratio ? ratio.turnover.toFixed(2) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
            label={t("daysOnHand")}
          >
            {ratio ? `${ratio.daysOnHand.toFixed(1)}d` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-4 bg-amber-500/10 border-amber-500/20"
            label={t("stockouts")}
          >
            {ratio ? String(ratio.stockoutCount) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-7">
        <CardHeader>
          <CardTitle className="text-sm">{t("healthMetrics")}</CardTitle>
          <CardDescription>{t("healthMetricsDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={ratioData}
            ariaLabel={t("healthTitle")}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-5">
        <CardHeader>
          <CardTitle className="text-sm">{t("flowTitle")}</CardTitle>
          <CardDescription>{t("turnoverVsDays")}</CardDescription>
        </CardHeader>
        <CardContent>
          <LineChart
            data={healthFlow}
            ariaLabel={t("healthTitle")}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
    </div>
  );
}
