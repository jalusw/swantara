"use client";

import { useTranslations } from "next-intl";
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
  const t = useTranslations("Dashboard");
  const payrollKpi = usePayrollKpi();
  const payroll = payrollKpi.data?.kpi;

  const payrollData = useMemo(
    () =>
      payroll
        ? [
            { label: t("grossCost"), value: payroll.grossCost },
            { label: t("netCost"), value: payroll.netCost },
          ]
        : [
            { label: t("grossCost"), value: 0 },
            { label: t("netCost"), value: 0 },
          ],
    [payroll, t],
  );

  const payrollComposition = useMemo(
    () =>
      payroll
        ? [
            { label: t("netCost"), value: payroll.netCost },
            {
              label: t("deductions"),
              value: Math.max(0, payroll.grossCost - payroll.netCost),
            },
          ]
        : [
            { label: t("netCost"), value: 0 },
            { label: t("deductions"), value: 0 },
          ],
    [payroll, t],
  );

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle>{t("payrollTitle")}</CardTitle>
          <CardDescription>{t("payrollDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <div className="col-span-12 sm:col-span-6 rounded-xl border border-border bg-primary/[0.04] p-5">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {t("grossCost")}
            </p>
            <p className="mt-2 font-heading text-2xl font-semibold tracking-tight tabular-nums">
              {payroll ? formatMoney(payroll.grossCost, { currency: DEFAULT_CURRENCY }) : "—"}
            </p>
          </div>
          <div className="col-span-12 sm:col-span-6 rounded-xl border border-border bg-card p-5">
            <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
              {t("netCost")}
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
            <CardTitle className="text-sm">{t("costComparison")}</CardTitle>
            <CardDescription>{t("grossVsNet")}</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <BarChart
              data={payrollData}
              ariaLabel={t("payrollTitle")}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            />
          </CardContent>
        </Card>
        <Card className="col-span-12 lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-sm">{t("netVsDeductions")}</CardTitle>
            <CardDescription>{t("grossComposition")}</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={payrollComposition}
              ariaLabel={t("payrollTitle")}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
              centerLabel={t("grossCost")}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function ProjectSection() {
  const t = useTranslations("Dashboard");
  const projectKpi = useProjectKpi();
  const project = projectKpi.data?.kpi;

  const projectData = useMemo(
    () =>
      project
        ? [
            { label: t("billed"), value: project.totalBilled },
            { label: t("margin"), value: project.totalMargin },
            { label: t("stat_expenses"), value: project.totalCost },
          ]
        : [
            { label: t("billed"), value: 0 },
            { label: t("margin"), value: 0 },
            { label: t("stat_expenses"), value: 0 },
          ],
    [project, t],
  );

  const projectComposition = useMemo(
    () =>
      project
        ? [
            { label: t("margin"), value: project.totalMargin },
            { label: t("stat_expenses"), value: project.totalCost },
          ]
        : [
            { label: t("margin"), value: 0 },
            { label: t("stat_expenses"), value: 0 },
          ],
    [project, t],
  );

  const utilizationData = useMemo(
    () =>
      project
        ? [
            { label: t("utilization"), value: project.utilization * 100 },
            { label: t("available"), value: (1 - project.utilization) * 100 },
          ]
        : [
            { label: t("utilization"), value: 0 },
            { label: t("available"), value: 0 },
          ],
    [project, t],
  );

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader className="pb-3">
          <CardTitle>{t("opsProjectsTitle")}</CardTitle>
          <CardDescription>{t("opsProjectsDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-5"
            label={t("opsProjectsTitle")}
          >
            {project ? String(project.projectCount) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 lg:col-span-4 rounded-xl border border-border bg-card p-5"
            label={t("utilization")}
          >
            {project ? `${(project.utilization * 100).toFixed(1)}%` : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-12 lg:col-span-4 rounded-xl border border-border bg-card p-5 bg-primary/[0.04] border-primary/15"
            label={t("billed")}
          >
            {project ? formatMoney(project.totalBilled, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-5"
            label={t("margin")}
          >
            {project ? formatMoney(project.totalMargin, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 rounded-xl border border-border bg-card p-5"
            label={t("stat_expenses")}
          >
            {project ? formatMoney(project.totalCost, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("billedVsMarginVsCost")}</CardTitle>
          <CardDescription>{t("financialBreakdown")}</CardDescription>
        </CardHeader>
        <CardContent className="pt-2">
          <BarChart
            data={projectData}
            ariaLabel={t("opsProjectsTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <div className="grid gap-6 grid-cols-12">
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">{t("marginVsCost")}</CardTitle>
            <CardDescription>{t("billedComposition")}</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={projectComposition}
              ariaLabel={t("billed")}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
              centerLabel={t("billed")}
            />
          </CardContent>
        </Card>
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">{t("utilization")}</CardTitle>
            <CardDescription>{t("usedCapacity")}</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={utilizationData}
              ariaLabel={t("utilization")}
              valueFormatter={(value) => `${value.toFixed(1)}%`}
              centerLabel={t("utilization")}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function SubscriptionSection() {
  const t = useTranslations("Dashboard");
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
            { label: t("churnRate"), value: sub.churnRate * 100 },
            { label: t("retained"), value: (1 - sub.churnRate) * 100 },
          ]
        : [
            { label: t("churnRate"), value: 0 },
            { label: t("retained"), value: 0 },
          ],
    [sub, t],
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
          <CardTitle>{t("opsSubscriptionsTitle")}</CardTitle>
          <CardDescription>{t("opsSubscriptionsDesc")}</CardDescription>
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
            label={t("churnRate")}
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
            label={t("churned")}
          >
            {sub ? String(sub.churned) : "—"}
          </KpiCard>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("recurringValue")}</CardTitle>
          <CardDescription>{t("mrrArrLtv")}</CardDescription>
        </CardHeader>
        <CardContent className="pt-2">
          <BarChart
            data={subscriptionData}
            ariaLabel={t("opsSubscriptionsTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <div className="grid gap-6 grid-cols-12">
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">{t("retention")}</CardTitle>
            <CardDescription>{t("churnVsRetained")}</CardDescription>
          </CardHeader>
          <CardContent className="pt-2">
            <DonutChart
              data={churnData}
              ariaLabel={t("churnRate")}
              valueFormatter={(value) => `${value.toFixed(1)}%`}
              centerLabel={t("churnRate")}
            />
          </CardContent>
        </Card>
        <Card className="col-span-12 sm:col-span-6">
          <CardHeader>
            <CardTitle className="text-sm">{t("growthTrajectory")}</CardTitle>
            <CardDescription>{t("recurringFlow")}</CardDescription>
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
