"use client";

import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { useProcurementKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatNumber } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function ProcurementKpiCard() {
  const { data } = useProcurementKpi();
  const kpi = data?.kpi;

  const procurementData = useMemo(
    () =>
      kpi
        ? [
            { label: "Purchase orders", value: kpi.purchaseCount },
            { label: "Avg cycle days", value: kpi.avgCycleDays },
            { label: "On-time", value: kpi.onTimeDeliveryPct * 100 },
            { label: "Price variance", value: kpi.priceVariancePct * 100 },
          ]
        : [
            { label: "Purchase orders", value: 0 },
            { label: "Avg cycle days", value: 0 },
            { label: "On-time", value: 0 },
            { label: "Price variance", value: 0 },
          ],
    [kpi],
  );

  const deliveryData = useMemo(
    () =>
      kpi
        ? [
            { label: "On-time", value: kpi.onTimeDeliveryPct * 100 },
            { label: "Late", value: (1 - kpi.onTimeDeliveryPct) * 100 },
          ]
        : [
            { label: "On-time", value: 0 },
            { label: "Late", value: 0 },
          ],
    [kpi],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Procurement"}</CardTitle>
          <CardDescription>{"Cycle time, on-time delivery and price variance."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4"
            label={"Purchase orders"}
          >
            {kpi ? String(kpi.purchaseCount) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4"
            label={"Avg cycle days"}
          >
            {kpi ? `${kpi.avgCycleDays.toFixed(1)}d` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={"On-time"}
          >
            {kpi ? `${(kpi.onTimeDeliveryPct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-4"
            label={"Price variance"}
          >
            {kpi ? `${(kpi.priceVariancePct * 100).toFixed(1)}%` : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-7">
        <CardHeader>
          <CardTitle className="text-sm">{"Procurement metrics"}</CardTitle>
          <CardDescription>{"Cycle, delivery and variance"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={procurementData}
            ariaLabel={"Procurement"}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-5">
        <CardHeader>
          <CardTitle className="text-sm">{"On-time delivery"}</CardTitle>
          <CardDescription>{"On-time vs late"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={deliveryData}
            ariaLabel={"On-time"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"On-time"}
          />
        </CardContent>
      </Card>
    </div>
  );
}
