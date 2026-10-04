"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useArApKpi, useCashKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";
import { KpiCard } from "./kpi-card";

export function CashSection() {
  const t = useTranslations("Dashboard");
  const cashKpi = useCashKpi();
  const arApKpi = useArApKpi();
  const cash = cashKpi.data?.kpi;
  const arAp = arApKpi.data?.kpi;

  const cashData = useMemo(
    () =>
      cash
        ? [
            { label: t("cashPosition"), value: cash.position },
            { label: t("cashBurn"), value: cash.burn },
            { label: t("cashForecast"), value: cash.forecast },
          ]
        : [
            { label: t("cashPosition"), value: 0 },
            { label: t("cashBurn"), value: 0 },
            { label: t("cashForecast"), value: 0 },
          ],
    [cash, t],
  );

  const arData = useMemo(
    () =>
      arAp
        ? [
            { label: t("overdueReceivables"), value: arAp.overdueArPct * 100 },
            { label: t("current"), value: (1 - arAp.overdueArPct) * 100 },
          ]
        : [
            { label: t("overdueReceivables"), value: 0 },
            { label: t("current"), value: 0 },
          ],
    [arAp, t],
  );

  const apData = useMemo(
    () =>
      arAp
        ? [
            { label: t("overduePayables"), value: arAp.overdueApPct * 100 },
            { label: t("current"), value: (1 - arAp.overdueApPct) * 100 },
          ]
        : [
            { label: t("overduePayables"), value: 0 },
            { label: t("current"), value: 0 },
          ],
    [arAp, t],
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
          <CardTitle>{t("cashTitle")}</CardTitle>
          <CardDescription>
            {t("cashDesc")} · {t("arApDesc")}
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 grid-cols-12">
          <KpiCard
            className="col-span-12 md:col-span-4 rounded-xl border border-border bg-card p-4 bg-primary/[0.04] border-primary/15"
            label={t("cashPosition")}
          >
            {cash ? formatMoney(cash.position, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 md:col-span-4 rounded-xl border border-border bg-card p-4"
            label={t("cashBurn")}
          >
            {cash ? formatMoney(cash.burn, { currency: DEFAULT_CURRENCY }) : "—"}
          </KpiCard>
          <KpiCard
            className="col-span-6 md:col-span-4 rounded-xl border border-border bg-card p-4"
            label={t("cashForecast")}
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
              {arAp
                ? t("overdueReceivablesPct", {
                    pct: (arAp.overdueArPct * 100).toFixed(1),
                  })
                : "—"}
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
              {arAp
                ? t("overduePayablesPct", {
                    pct: (arAp.overdueApPct * 100).toFixed(1),
                  })
                : "—"}
            </p>
          </div>
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-8">
        <CardHeader>
          <CardTitle className="text-sm">{t("cashFlowTitle")}</CardTitle>
          <CardDescription>{t("cashFlowDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={cashData}
            ariaLabel={t("cashTitle")}
            valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
          />
        </CardContent>
      </Card>
      <Card className="col-span-12 lg:col-span-4">
        <CardHeader>
          <CardTitle className="text-sm">{"DSO / DPO"}</CardTitle>
          <CardDescription>{t("dsoDpoDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <BarChart
            data={arApComparison}
            ariaLabel={t("arApTitle")}
            valueFormatter={(value) => `${value.toFixed(1)}d`}
          />
        </CardContent>
      </Card>
      <Card className="col-span-6">
        <CardHeader>
          <CardTitle className="text-sm">{"DSO"}</CardTitle>
          <CardDescription>{t("overdueVsCurrent")}</CardDescription>
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
          <CardDescription>{t("overdueVsCurrent")}</CardDescription>
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
