"use client";

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
  const salesKpi = useSalesKpi();
  const sales = salesKpi.data?.kpi;

  const salesData = useMemo(
    () =>
      sales
        ? [
            { label: "Bookings", value: sales.bookings },
            { label: "Revenue", value: sales.revenue },
            { label: "COGS", value: sales.cogs },
            { label: "Gross margin", value: sales.grossMargin },
          ]
        : [
            { label: "Bookings", value: 0 },
            { label: "Revenue", value: 0 },
            { label: "COGS", value: 0 },
            { label: "Gross margin", value: 0 },
          ],
    [sales],
  );

  const winRateData = useMemo(
    () =>
      sales
        ? [
            { label: "Win rate", value: sales.winRate * 100 },
            { label: "Loss", value: (1 - sales.winRate) * 100 },
          ]
        : [
            { label: "Win rate", value: 0 },
            { label: "Loss", value: 0 },
          ],
    [sales],
  );

  const marginFlow = useMemo(
    () =>
      sales
        ? [
            { label: "Bookings", value: sales.bookings },
            { label: "Revenue", value: sales.revenue },
            { label: "Gross margin", value: sales.grossMargin },
          ]
        : [
            { label: "Bookings", value: 0 },
            { label: "Revenue", value: 0 },
            { label: "Gross margin", value: 0 },
          ],
    [sales],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Sales"}</CardTitle>
          <CardDescription>{"Bookings, revenue, COGS and win rate."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"Bookings"}
          >
            {sales ? formatMoney(sales.bookings, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"Revenue"}
          >
            {sales ? formatMoney(sales.revenue, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"COGS"}
          >
            {sales ? formatMoney(sales.cogs, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <div className="col-span-6 rounded-xl border border-border bg-card p-4">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {"Gross margin"}
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
            label={"Win rate"}
          >
            {sales ? `${(sales.winRate * 100).toFixed(1)}%` : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{"Bookings flow"}</CardTitle>
          <CardDescription>{"Bookings → Revenue → Margin"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={salesData}
            ariaLabel={"Sales"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"Win rate"}</CardTitle>
          <CardDescription>{"Win vs loss"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={winRateData}
            ariaLabel={"Win rate"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"Win rate"}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"Margin trend"}</CardTitle>
          <CardDescription>{"Flow across stages"}</CardDescription>
        </CardHeader>
        <CardContent>
          <LineChart
            data={marginFlow}
            ariaLabel={"Gross margin"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export function PipelineSection() {
  const pipelineKpi = usePipelineKpi();
  const pipeline = pipelineKpi.data?.kpi;

  const pipelineData = useMemo(
    () =>
      pipeline
        ? [
            { label: "Weighted pipeline", value: pipeline.weightedPipeline },
            { label: "Expected revenue", value: pipeline.totalExpectedRevenue },
          ]
        : [
            { label: "Weighted pipeline", value: 0 },
            { label: "Expected revenue", value: 0 },
          ],
    [pipeline],
  );

  const pipelineWinData = useMemo(
    () =>
      pipeline
        ? [
            { label: "Win rate", value: pipeline.winRate * 100 },
            { label: "Loss", value: (1 - pipeline.winRate) * 100 },
          ]
        : [
            { label: "Win rate", value: 0 },
            { label: "Loss", value: 0 },
          ],
    [pipeline],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Pipeline"}</CardTitle>
          <CardDescription>{"Weighted pipeline and stage coverage."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={"Weighted pipeline"}
          >
            {pipeline
              ? formatMoney(pipeline.weightedPipeline, { currency: DEFAULT_CURRENCY })
              : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"Expected revenue"}
          >
            {pipeline
              ? formatMoney(pipeline.totalExpectedRevenue, { currency: DEFAULT_CURRENCY })
              : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"Win rate"}
          >
            {pipeline ? `${(pipeline.winRate * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"Stages"}
          >
            {pipeline ? String(pipeline.stages) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{"Pipeline value"}</CardTitle>
          <CardDescription>{"Weighted vs expected"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={pipelineData}
            ariaLabel={"Pipeline"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{"Win rate"}</CardTitle>
          <CardDescription>{"Coverage across stages"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={pipelineWinData}
            ariaLabel={"Win rate"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"Win rate"}
          />
        </CardContent>
      </Card>
    </div>
  );
}
