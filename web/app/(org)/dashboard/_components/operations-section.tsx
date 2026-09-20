"use client";

import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { LineChart } from "@/components/line-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { usePayrollKpi, useProjectKpi, useSubscriptionKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function PayrollSection() {
  const payrollKpi = usePayrollKpi();
  const payroll = payrollKpi.data?.kpi;

  const payrollData = useMemo(
    () =>
      payroll
        ? [
            { label: "Gross cost", value: payroll.grossCost },
            { label: "Net cost", value: payroll.netCost },
          ]
        : [
            { label: "Gross cost", value: 0 },
            { label: "Net cost", value: 0 },
          ],
    [payroll],
  );

  const payrollComposition = useMemo(
    () =>
      payroll
        ? [
            { label: "Net cost", value: payroll.netCost },
            {
              label: "Deductions",
              value: Math.max(0, payroll.grossCost - payroll.netCost),
            },
          ]
        : [
            { label: "Net cost", value: 0 },
            { label: "Deductions", value: 0 },
          ],
    [payroll],
  );

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle>{"Payroll"}</CardTitle>
          <CardDescription>{"Gross and net payroll cost."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <div className="col-span-12 sm:col-span-6 rounded-xl border border-border bg-primary/[0.04] p-5">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {"Gross cost"}
            </p>
            <p className="mt-2 font-heading text-2xl font-semibold tracking-tight tabular-nums">
              {payroll ? formatMoney(payroll.grossCost, { currency: DEFAULT_CURRENCY }) : "—"}
            </p>
          </div>
          <div className="col-span-12 sm:col-span-6 rounded-xl border border-border bg-card p-5">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {"Net cost"}
            </p>
            <p className="mt-2 font-heading text-2xl font-semibold tracking-tight tabular-nums">
              {payroll ? formatMoney(payroll.netCost, { currency: DEFAULT_CURRENCY }) : "—"}
            </p>
          </div>
        </CardContent>
      </Card>
      <div className="grid gap-6 grid-cols-12">
        <Card className="col-span-12 lg:col-span-8">
          <CardHeader>
            <CardTitle className="text-sm">Cost comparison</CardTitle>
            <CardDescription>Gross vs net</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <BarChart
              data={payrollData}
              ariaLabel={"Payroll"}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            />
          </CardContent>
        </Card>
        <Card className="col-span-12 lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-sm">Net vs deductions</CardTitle>
            <CardDescription>Composition of gross</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={payrollComposition}
              ariaLabel={"Payroll"}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
              centerLabel={"Gross cost"}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function ProjectSection() {
  const projectKpi = useProjectKpi();
  const project = projectKpi.data?.kpi;

  const projectData = useMemo(
    () =>
      project
        ? [
            { label: "Billed", value: project.totalBilled },
            { label: "Margin", value: project.totalMargin },
            { label: "Cost", value: project.totalCost },
          ]
        : [
            { label: "Billed", value: 0 },
            { label: "Margin", value: 0 },
            { label: "Cost", value: 0 },
          ],
    [project],
  );

  const projectComposition = useMemo(
    () =>
      project
        ? [
            { label: "Margin", value: project.totalMargin },
            { label: "Cost", value: project.totalCost },
          ]
        : [
            { label: "Margin", value: 0 },
            { label: "Cost", value: 0 },
          ],
    [project],
  );

  const utilizationData = useMemo(
    () =>
      project
        ? [
            { label: "Utilization", value: project.utilization * 100 },
            { label: "Available", value: (1 - project.utilization) * 100 },
          ]
        : [
            { label: "Utilization", value: 0 },
            { label: "Available", value: 0 },
          ],
    [project],
  );

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle>{"Projects"}</CardTitle>
          <CardDescription>{"Margin, cost, billed and utilization."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-5"
            label={"Projects"}
          >
            {project ? String(project.projectCount) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-5"
            label={"Utilization"}
          >
            {project ? `${(project.utilization * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-5 bg-primary/[0.04] border-primary/15"
            label={"Billed"}
          >
            {project ? formatMoney(project.totalBilled, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-5"
            label={"Margin"}
          >
            {project ? formatMoney(project.totalMargin, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-5"
            label={"Cost"}
          >
            {project ? formatMoney(project.totalCost, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Billed vs margin vs cost</CardTitle>
          <CardDescription>Financial split</CardDescription>
        </CardHeader>
        <CardContent className="pt-2">
          <BarChart
            data={projectData}
            ariaLabel={"Projects"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <div className="grid gap-6 grid-cols-12">
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">Margin vs cost</CardTitle>
            <CardDescription>Billed composition</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={projectComposition}
              ariaLabel={"Billed"}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
              centerLabel={"Billed"}
            />
          </CardContent>
        </Card>
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">Utilization</CardTitle>
            <CardDescription>Capacity used</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={utilizationData}
              ariaLabel={"Utilization"}
              valueFormatter={(value) => `${value.toFixed(1)}%`}
              centerLabel={"Utilization"}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function SubscriptionSection() {
  const subscriptionKpi = useSubscriptionKpi();
  const sub = subscriptionKpi.data?.kpi;

  const subscriptionData = useMemo(
    () =>
      sub
        ? [
            { label: "MRR", value: sub.mrr },
            { label: "ARR", value: sub.arr },
            { label: "LTV", value: sub.ltv },
          ]
        : [
            { label: "MRR", value: 0 },
            { label: "ARR", value: 0 },
            { label: "LTV", value: 0 },
          ],
    [sub],
  );

  const churnData = useMemo(
    () =>
      sub
        ? [
            { label: "Churn rate", value: sub.churnRate * 100 },
            { label: "Retained", value: (1 - sub.churnRate) * 100 },
          ]
        : [
            { label: "Churn rate", value: 0 },
            { label: "Retained", value: 0 },
          ],
    [sub],
  );

  const recurringFlow = useMemo(
    () =>
      sub
        ? [
            { label: "MRR", value: sub.mrr },
            { label: "ARR", value: sub.arr },
            { label: "LTV", value: sub.ltv },
          ]
        : [
            { label: "MRR", value: 0 },
            { label: "ARR", value: 0 },
            { label: "LTV", value: 0 },
          ],
    [sub],
  );

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle>{"Subscriptions"}</CardTitle>
          <CardDescription>{"MRR, ARR, churn and lifetime value."}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-5"
            label={"MRR"}
          >
            {sub ? formatMoney(sub.mrr, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-5"
            label={"ARR"}
          >
            {sub ? formatMoney(sub.arr, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-5"
            label={"Churn rate"}
          >
            {sub ? `${(sub.churnRate * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-3 rounded-xl border border-border bg-card p-5"
            label={"LTV"}
          >
            {sub ? formatMoney(sub.ltv, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-12 rounded-xl border border-amber-500/20 bg-amber-500/10 p-5"
            label={"Churned"}
          >
            {sub ? String(sub.churned) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Recurring value</CardTitle>
          <CardDescription>MRR, ARR and LTV</CardDescription>
        </CardHeader>
        <CardContent className="pt-2">
          <BarChart
            data={subscriptionData}
            ariaLabel={"Subscriptions"}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <div className="grid gap-6 grid-cols-12">
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">Retention</CardTitle>
            <CardDescription>Churn vs retained</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={churnData}
              ariaLabel={"Churn rate"}
              valueFormatter={(value) => `${value.toFixed(1)}%`}
              centerLabel={"Churn rate"}
            />
          </CardContent>
        </Card>
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">Growth trajectory</CardTitle>
            <CardDescription>Recurring flow</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <LineChart
              data={recurringFlow}
              ariaLabel={"MRR"}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
