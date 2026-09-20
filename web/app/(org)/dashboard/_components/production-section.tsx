"use client";

import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { useManufacturingKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatNumber } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function ManufacturingSection() {
  const { data } = useManufacturingKpi();
  const kpi = data?.kpi;

  const manufacturingData = useMemo(
    () =>
      kpi
        ? [
            { label: "Orders", value: kpi.orderCount },
            { label: "OEE", value: kpi.oeePct * 100 },
            { label: "Yield", value: kpi.yieldPct * 100 },
            { label: "Scrap", value: kpi.scrapPct * 100 },
            { label: "Cost variance", value: kpi.costVariancePct * 100 },
          ]
        : [
            { label: "Orders", value: 0 },
            { label: "OEE", value: 0 },
            { label: "Yield", value: 0 },
            { label: "Scrap", value: 0 },
            { label: "Cost variance", value: 0 },
          ],
    [kpi],
  );

  const yieldScrapData = useMemo(
    () =>
      kpi
        ? [
            { label: "Yield", value: kpi.yieldPct * 100 },
            { label: "Scrap", value: kpi.scrapPct * 100 },
          ]
        : [
            { label: "Yield", value: 0 },
            { label: "Scrap", value: 0 },
          ],
    [kpi],
  );

  const oeeVarianceData = useMemo(
    () =>
      kpi
        ? [
            { label: "OEE", value: kpi.oeePct * 100 },
            { label: "Cost variance", value: kpi.costVariancePct * 100 },
          ]
        : [
            { label: "OEE", value: 0 },
            { label: "Cost variance", value: 0 },
          ],
    [kpi],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Manufacturing"}</CardTitle>
          <CardDescription>{"OEE, yield, scrap and cost variance."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={"Orders"}
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
            label={"Yield"}
          >
            {kpi ? `${(kpi.yieldPct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-2 rounded-xl border border-border bg-card p-4"
            label={"Scrap"}
          >
            {kpi ? `${(kpi.scrapPct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-2 rounded-xl border border-border bg-card p-4"
            label={"Cost variance"}
          >
            {kpi ? `${(kpi.costVariancePct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{"Output metrics"}</CardTitle>
          <CardDescription>{"Orders, OEE, yield and variance"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={manufacturingData}
            ariaLabel={"Manufacturing"}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"Yield vs scrap"}</CardTitle>
          <CardDescription>{"Quality balance"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={yieldScrapData}
            ariaLabel={"Yield"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"Yield"}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"OEE vs variance"}</CardTitle>
          <CardDescription>{"Efficiency vs cost"}</CardDescription>
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
