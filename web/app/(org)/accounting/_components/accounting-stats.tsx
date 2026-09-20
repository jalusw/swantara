"use client";

import { Banknote, Receipt, TrendingUp, Wallet } from "lucide-react";

import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useCashKpi, useFinanceKpi } from "@/lib/hooks/use-dashboard-kpis";
import { formatMoney } from "@/lib/utils";

export function AccountingStats() {
  const { data: financeData } = useFinanceKpi();
  const { data: cashData } = useCashKpi();

  const kpi = financeData?.kpi;
  const cashKpi = cashData?.kpi;

  return (
    <MetricGrid>
      <StatCard
        label={"Revenue"}
        value={kpi ? formatMoney(kpi.revenue, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={Banknote}
      />
      <StatCard
        label={"Expenses"}
        value={kpi ? formatMoney(kpi.expenses, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={Receipt}
      />
      <StatCard
        label={"Net profit"}
        value={kpi ? formatMoney(kpi.ebitda, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={TrendingUp}
        trend={kpi ? `${(kpi.netMarginPct * 100).toFixed(0)}% from last month` : undefined}
      />
      <StatCard
        label={"Cash balance"}
        value={cashKpi ? formatMoney(cashKpi.position, { currency: DEFAULT_CURRENCY }) : "—"}
        icon={Wallet}
        trend={
          cashKpi
            ? `${formatMoney(cashKpi.forecast, { currency: DEFAULT_CURRENCY })} from last month`
            : undefined
        }
        trendDirection={cashKpi && cashKpi.burn > 0 ? "down" : undefined}
      />
    </MetricGrid>
  );
}
