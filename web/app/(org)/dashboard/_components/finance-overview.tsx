"use client";

import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useFinanceKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function FinanceOverview() {
  const financeKpi = useFinanceKpi();
  const kpi = financeKpi.data?.kpi;

  const financeData = useMemo(
    () =>
      kpi
        ? [
            { label: "Revenue", value: kpi.revenue },
            { label: "Expenses", value: kpi.expenses },
            { label: "EBITDA", value: kpi.ebitda },
          ]
        : [
            { label: "Revenue", value: 0 },
            { label: "Expenses", value: 0 },
            { label: "EBITDA", value: 0 },
          ],
    [kpi],
  );

  const compositionData = useMemo(
    () =>
      kpi
        ? [
            { label: "Expenses", value: kpi.expenses },
            { label: "EBITDA", value: kpi.ebitda },
          ]
        : [
            { label: "Expenses", value: 0 },
            { label: "EBITDA", value: 0 },
          ],
    [kpi],
  );

  const marginData = useMemo(
    () =>
      kpi
        ? [
            { label: "Gross margin", value: kpi.grossMarginPct * 100 },
            { label: "Net margin", value: kpi.netMarginPct * 100 },
          ]
        : [
            { label: "Gross margin", value: 0 },
            { label: "Net margin", value: 0 },
          ],
    [kpi],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12 lg:col-span-8">
        <CardHeader>
          <CardTitle>{"Finance"}</CardTitle>
          <CardDescription>{"General ledger health and margins."}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={financeData}
            ariaLabel={"Finance"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-4 bg-muted/20">
        <CardHeader>
          <CardTitle className="text-sm">{"Composition"}</CardTitle>
          <CardDescription>{"Expenses vs EBITDA"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={compositionData}
            ariaLabel={"Finance"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            centerLabel={"Revenue"}
          />
        </CardContent>
      </Card>
      <div className="col-span-12 grid gap-4 grid-cols-12">
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={"Revenue"}
        >
          {kpi ? formatMoney(kpi.revenue, { currency: DEFAULT_CURRENCY }) : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={"Expenses"}
        >
          {kpi ? formatMoney(kpi.expenses, { currency: DEFAULT_CURRENCY }) : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
          label={"EBITDA"}
        >
          {kpi ? formatMoney(kpi.ebitda, { currency: DEFAULT_CURRENCY }) : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={"Gross margin"}
        >
          {kpi ? `${(kpi.grossMarginPct * 100).toFixed(1)}%` : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={"Net margin"}
        >
          {kpi ? `${(kpi.netMarginPct * 100).toFixed(1)}%` : "—"}
        </KpiCard>
        <KpiCard
          className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-4"
          label={"Current ratio"}
        >
          {kpi ? kpi.currentRatio.toFixed(2) : "—"}
        </KpiCard>
      </div>
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle className="text-sm">{"Margins"}</CardTitle>
          <CardDescription>{"Gross vs Net"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={marginData}
            ariaLabel={"Gross margin"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
          />
        </CardContent>
      </Card>
    </div>
  );
}
