"use client";

import { Banknote, Receipt, TrendingUp, Wallet } from "lucide-react";
import { useTranslations } from "next-intl";

import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useCashKpi, useFinanceKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function AccountingStats() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const { data: financeData } = useFinanceKpi();
  const { data: cashData } = useCashKpi();

  const kpi = financeData?.kpi;
  const cashKpi = cashData?.kpi;

  return (
    <MetricGrid>
      <StatCard
        label={t("statRevenue")}
        value={kpi ? formatMoney(kpi.revenue, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={Banknote}
      />
      <StatCard
        label={t("statExpenses")}
        value={kpi ? formatMoney(kpi.expenses, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={Receipt}
      />
      <StatCard
        label={t("statNetProfit")}
        value={kpi ? formatMoney(kpi.ebitda, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={TrendingUp}
        trend={kpi ? t("trendMargin", { value: (kpi.netMarginPct * 100).toFixed(0) }) : undefined}
      />
      <StatCard
        label={t("statCashBalance")}
        value={cashKpi ? formatMoney(cashKpi.position, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={Wallet}
        trend={
          cashKpi
            ? t("trendForecast", {
                value: formatMoney(cashKpi.forecast, { currency: DEFAULT_CURRENCY }),
              })
            : undefined
        }
        trendDirection={cashKpi && cashKpi.burn > 0 ? "down" : undefined}
      />
    </MetricGrid>
  );
}
