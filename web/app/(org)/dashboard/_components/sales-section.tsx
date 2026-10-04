"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { LineChart } from "@/components/line-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { usePipelineKpi, useSalesKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function SalesSection() {
  const t = useTranslations("Dashboard");
  const salesKpi = useSalesKpi();
  const sales = salesKpi.data?.kpi;

  const salesData = useMemo(
    () =>
      sales
        ? [
            { label: t("bookings"), value: sales.bookings },
            { label: t("stat_revenue"), value: sales.revenue },
            { label: t("cogs"), value: sales.cogs },
            { label: t("grossMargin"), value: sales.grossMargin },
          ]
        : [
            { label: t("bookings"), value: 0 },
            { label: t("stat_revenue"), value: 0 },
            { label: t("cogs"), value: 0 },
            { label: t("grossMargin"), value: 0 },
          ],
    [sales, t],
  );

  const winRateData = useMemo(
    () =>
      sales
        ? [
            { label: t("winRate"), value: sales.winRate * 100 },
            { label: t("loss"), value: (1 - sales.winRate) * 100 },
          ]
        : [
            { label: t("winRate"), value: 0 },
            { label: t("loss"), value: 0 },
          ],
    [sales, t],
  );

  const marginFlow = useMemo(
    () =>
      sales
        ? [
            { label: t("bookings"), value: sales.bookings },
            { label: t("stat_revenue"), value: sales.revenue },
            { label: t("grossMargin"), value: sales.grossMargin },
          ]
        : [
            { label: t("bookings"), value: 0 },
            { label: t("stat_revenue"), value: 0 },
            { label: t("grossMargin"), value: 0 },
          ],
    [sales, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{t("salesTitle")}</CardTitle>
          <CardDescription>{t("salesDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("bookings")}
          >
            {sales ? formatMoney(sales.bookings, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("stat_revenue")}
          >
            {sales ? formatMoney(sales.revenue, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("cogs")}
          >
            {sales ? formatMoney(sales.cogs, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <div className="col-span-6 rounded-xl border border-border bg-card p-4">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {t("grossMargin")}
            </p>
            <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
              {sales ? formatMoney(sales.grossMargin, { currency: DEFAULT_CURRENCY }) : "—"}
            </p>
            <p className="text-xs text-muted-foreground">
              {sales ? `${(sales.grossMarginPct * 100).toFixed(1)}%` : "—"}
            </p>
          </div>
          <KpiCard
            className="col-span-12 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={t("winRate")}
          >
            {sales ? `${(sales.winRate * 100).toFixed(1)}%` : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{t("bookingFlowTitle")}</CardTitle>
          <CardDescription>{t("bookingFlowDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={salesData}
            ariaLabel={t("salesTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{t("winRate")}</CardTitle>
          <CardDescription>{t("winVsLoss")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={winRateData}
            ariaLabel={t("winRate")}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={t("winRate")}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{t("marginTrendTitle")}</CardTitle>
          <CardDescription>{t("marginFlowDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <LineChart
            data={marginFlow}
            ariaLabel={t("grossMargin")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export function PipelineSection() {
  const t = useTranslations("Dashboard");
  const pipelineKpi = usePipelineKpi();
  const pipeline = pipelineKpi.data?.kpi;

  const pipelineData = useMemo(
    () =>
      pipeline
        ? [
            { label: t("weightedPipeline"), value: pipeline.weightedPipeline },
            { label: t("expectedRevenue"), value: pipeline.totalExpectedRevenue },
          ]
        : [
            { label: t("weightedPipeline"), value: 0 },
            { label: t("expectedRevenue"), value: 0 },
          ],
    [pipeline, t],
  );

  const pipelineWinData = useMemo(
    () =>
      pipeline
        ? [
            { label: t("winRate"), value: pipeline.winRate * 100 },
            { label: t("loss"), value: (1 - pipeline.winRate) * 100 },
          ]
        : [
            { label: t("winRate"), value: 0 },
            { label: t("loss"), value: 0 },
          ],
    [pipeline, t],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{t("pipelineTitle")}</CardTitle>
          <CardDescription>{t("pipelineDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={t("weightedPipeline")}
          >
            {pipeline
              ? formatMoney(pipeline.weightedPipeline, { currency: DEFAULT_CURRENCY })
              : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("expectedRevenue")}
          >
            {pipeline
              ? formatMoney(pipeline.totalExpectedRevenue, { currency: DEFAULT_CURRENCY })
              : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("winRate")}
          >
            {pipeline ? `${(pipeline.winRate * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={t("stages")}
          >
            {pipeline ? String(pipeline.stages) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{t("pipelineValueTitle")}</CardTitle>
          <CardDescription>{t("pipelineValueDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={pipelineData}
            ariaLabel={t("pipelineTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{t("winRate")}</CardTitle>
          <CardDescription>{t("pipelineWinDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={pipelineWinData}
            ariaLabel={t("winRate")}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={t("winRate")}
          />
        </CardContent>
      </Card>
    </div>
  );
}
