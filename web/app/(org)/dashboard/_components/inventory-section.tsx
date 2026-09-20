"use client";

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
  const inventoryKpi = useInventoryKpi();
  const inv = inventoryKpi.data?.kpi;

  const inventoryData = useMemo(
    () =>
      inv
        ? [
            { label: "On-hand value", value: inv.onHandValue },
            { label: "On-hand qty", value: inv.onHandQuantity },
          ]
        : [
            { label: "On-hand value", value: 0 },
            { label: "On-hand qty", value: 0 },
          ],
    [inv],
  );

  const stockComposition = useMemo(
    () =>
      inv
        ? [
            { label: "On-hand qty", value: inv.onHandQuantity },
            { label: "Products", value: inv.productCount },
          ]
        : [
            { label: "On-hand qty", value: 0 },
            { label: "Products", value: 0 },
          ],
    [inv],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Inventory"}</CardTitle>
          <CardDescription>{"On-hand value, quantity and catalog breadth."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-12 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={"On-hand value"}
          >
            {inv ? formatMoney(inv.onHandValue, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"On-hand qty"}
          >
            {inv ? formatNumber(inv.onHandQuantity) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-4"
            label={"Products"}
          >
            {inv ? formatNumber(inv.productCount) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-7">
        <CardHeader>
          <CardTitle className="text-sm">{"On-hand breakdown"}</CardTitle>
          <CardDescription>{"Value vs quantity"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={inventoryData}
            ariaLabel={"Inventory"}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-5">
        <CardHeader>
          <CardTitle className="text-sm">{"Catalog breadth"}</CardTitle>
          <CardDescription>{"Quantity vs products"}</CardDescription>
        </CardHeader>
        <CardContent>
          <DonutChart
            data={stockComposition}
            ariaLabel={"Inventory"}
            valueFormatter={(value) => formatNumber(value)}
            centerLabel={"Products"}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export function InventoryHealthSection() {
  const ratioKpi = useInventoryRatioKpi();
  const ratio = ratioKpi.data?.kpi;

  const ratioData = useMemo(
    () =>
      ratio
        ? [
            { label: "Turnover", value: ratio.turnover },
            { label: "Days on hand", value: ratio.daysOnHand },
            { label: "Stockouts", value: ratio.stockoutCount },
          ]
        : [
            { label: "Turnover", value: 0 },
            { label: "Days on hand", value: 0 },
            { label: "Stockouts", value: 0 },
          ],
    [ratio],
  );

  const healthFlow = useMemo(
    () =>
      ratio
        ? [
            { label: "Turnover", value: ratio.turnover },
            { label: "Days on hand", value: ratio.daysOnHand },
          ]
        : [
            { label: "Turnover", value: 0 },
            { label: "Days on hand", value: 0 },
          ],
    [ratio],
  );

  return (
    <div className="grid gap-4 grid-cols-12">
      <Card className="col-span-12">
        <CardHeader>
          <CardTitle>{"Inventory health"}</CardTitle>
          <CardDescription>{"Turnover, days on hand and stockouts."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
            label={"Turnover"}
          >
            {ratio ? ratio.turnover.toFixed(2) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-4"
            label={"Days on hand"}
          >
            {ratio ? `${ratio.daysOnHand.toFixed(1)}d` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-4 bg-amber-500/10 border-amber-500/20"
            label={"Stockouts"}
          >
            {ratio ? String(ratio.stockoutCount) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-7">
        <CardHeader>
          <CardTitle className="text-sm">{"Health metrics"}</CardTitle>
          <CardDescription>{"Turnover, days and stockouts"}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={ratioData}
            ariaLabel={"Inventory health"}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-5">
        <CardHeader>
          <CardTitle className="text-sm">{"Flow"}</CardTitle>
          <CardDescription>{"Turnover vs days on hand"}</CardDescription>
        </CardHeader>
        <CardContent>
          <LineChart
            data={healthFlow}
            ariaLabel={"Inventory health"}
            valueFormatter={(value) => formatNumber(value)}
          />
        </CardContent>
      </Card>
    </div>
  );
}
