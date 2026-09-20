"use client";

import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useArApKpi, useCashKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function CashSection() {
  const cashKpi = useCashKpi();
  const arApKpi = useArApKpi();
  const cash = cashKpi.data?.kpi;
  const arAp = arApKpi.data?.kpi;

  const cashData = useMemo(
    () =>
      cash
        ? [
            { label: "Cash position", value: cash.position },
            { label: "Burn", value: cash.burn },
            { label: "Forecast", value: cash.forecast },
          ]
        : [
            { label: "Cash position", value: 0 },
            { label: "Burn", value: 0 },
            { label: "Forecast", value: 0 },
          ],
    [cash],
  );

  const arData = useMemo(
    () =>
      arAp
        ? [
            { label: "Overdue AR", value: arAp.overdueArPct * 100 },
            { label: "Current", value: (1 - arAp.overdueArPct) * 100 },
          ]
        : [
            { label: "Overdue AR", value: 0 },
            { label: "Current", value: 0 },
          ],
    [arAp],
  );

  const apData = useMemo(
    () =>
      arAp
        ? [
            { label: "Overdue AP", value: arAp.overdueApPct * 100 },
            { label: "Current", value: (1 - arAp.overdueApPct) * 100 },
          ]
        : [
            { label: "Overdue AP", value: 0 },
            { label: "Current", value: 0 },
          ],
    [arAp],
  );

  const arApComparison = useMemo(
    () =>
      arAp
        ? [
            { label: "DSO", value: arAp.dso },
            { label: "DPO", value: arAp.dpo },
          ]
        : [
            { label: "DSO", value: 0 },
            { label: "DPO", value: 0 },
          ],
    [arAp],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Cash"}</CardTitle>
          <CardDescription>
            {"Bank position, burn and forecast."} · {"DSO, DPO and overdue exposure."}
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-12 md:col-span-4 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={"Cash position"}
          >
            {cash ? formatMoney(cash.position, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 md:col-span-4 rounded-xl border border-border bg-card p-4"
            label={"Burn"}
          >
            {cash ? formatMoney(cash.burn, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 md:col-span-4 rounded-xl border border-border bg-card p-4"
            label={"Forecast"}
          >
            {cash ? formatMoney(cash.forecast, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <div className="col-span-6 rounded-xl border border-border p-4">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {"DSO"}
            </p>
            <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
              {arAp ? `${arAp.dso.toFixed(1)}d` : "—"}
            </p>
            <p className="text-xs text-muted-foreground">
              {arAp ? `${(arAp.overdueArPct * 100).toFixed(1)}% ${"Overdue AR"}` : "—"}
            </p>
          </div>
          <div className="col-span-6 rounded-xl border border-border p-4">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {"DPO"}
            </p>
            <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
              {arAp ? `${arAp.dpo.toFixed(1)}d` : "—"}
            </p>
            <p className="text-xs text-muted-foreground">
              {arAp ? `${(arAp.overdueApPct * 100).toFixed(1)}% ${"Overdue AP"}` : "—"}
            </p>
          </div>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-8">
        <CardHeader>
          <CardTitle className="text-sm">{"Cash flow"}</CardTitle>
          <CardDescription>{"Position, burn and forecast"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={cashData}
            ariaLabel={"Cash"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-4">
        <CardHeader>
          <CardTitle className="text-sm">{"DSO / DPO"}</CardTitle>
          <CardDescription>{"Collection vs payment days"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={arApComparison}
            ariaLabel={"Receivables & Payables"}
            valueFormatter={(value) => `${value.toFixed(1)}d`}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"DSO"}</CardTitle>
          <CardDescription>{"Overdue vs current"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={arData}
            ariaLabel={"DSO"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"DSO"}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"DPO"}</CardTitle>
          <CardDescription>{"Overdue vs current"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={apData}
            ariaLabel={"DPO"}
            valueFormatter={(value) => `${value.toFixed(1)}%`}
            centerLabel={"DPO"}
          />
        </CardContent>
      </Card>
    </div>
  );
}
