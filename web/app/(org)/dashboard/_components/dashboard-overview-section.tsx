"use client";

import { Banknote, BarChart3, Truck, UserRound, Users } from "lucide-react";
import { useTranslations } from "next-intl";
import { useMemo } from "react";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { DonutChart } from "@/components/donut-chart";
import { PageHeader } from "@/components/page-header";
import { Separator } from "@/components/separator";
import { StatCard } from "@/components/stat-card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useFinanceKpi, useInventoryKpi, useProcurementKpi } from "@/lib/hooks/use-dashboard-kpis";
import { useMeOrganizationsQuery, useMeQuery } from "@/lib/hooks/use-me-query";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import { usePermissions } from "@/lib/hooks/use-permissions";
import type { Contact, Employee } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatMoney, formatNumber } from "@/lib/utils";
import { DashboardActions } from "./dashboard-actions";
import { ModulesGrid } from "./modules-grid";

type DashboardStat = {
  key: "revenue" | "customers" | "suppliers" | "employees";
  value: string;
  trend: string;
  icon: typeof Banknote;
  trendDirection?: "up" | "down";
  permission: string;
};

function useDashboardStats(): DashboardStat[] {
  const { has } = usePermissions();
  const t = useTranslations("Dashboard");
  const financeKpi = useFinanceKpi();
  const inventoryKpi = useInventoryKpi();
  const procurementKpi = useProcurementKpi();
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>, number>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
    {},
    { select: (data) => data.contacts.length },
  );
  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>, number>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
    {},
    { select: (data) => data.employees.length },
  );

  const contactsCount = contactsQuery.data ?? 0;
  const employeesCount = employeesQuery.data ?? 0;
  const finance = financeKpi.data?.kpi;
  const inventory = inventoryKpi.data?.kpi;
  const procurement = procurementKpi.data?.kpi;

  const stats: DashboardStat[] = [
    {
      key: "revenue",
      value: formatMoney(finance?.revenue ?? 0, { currency: DEFAULT_CURRENCY }),
      trend: finance?.netMarginPct
        ? t("trendMargin", {
            pct: (finance.netMarginPct * 100).toFixed(0),
          })
        : "—",
      icon: Banknote,
      permission: "journal_entry.view",
    },
    {
      key: "customers",
      value: formatNumber(contactsCount),
      trend: t("trendFlat"),
      icon: Users,
      permission: "contact.view",
    },
    {
      key: "suppliers",
      value: formatNumber(inventory?.productCount ?? contactsCount),
      trend: t("trendFlat"),
      icon: Truck,
      permission: "contact.view",
    },
    {
      key: "employees",
      value: formatNumber(employeesCount || procurement?.purchaseCount || 0),
      trend: t("trendFlat"),
      icon: UserRound,
      permission: "employee.view",
    },
  ];

  return stats.filter((stat) => has(stat.permission));
}

export function DashboardOverview({ orgId }: { orgId: string }) {
  const t = useTranslations("Dashboard");
  const statLabel = (key: string) => (t as unknown as (k: string) => string)(`stat_${key}`);
  const { data: me } = useMeQuery();
  const { data: orgList } = useMeOrganizationsQuery();
  const financeKpi = useFinanceKpi();
  const kpi = financeKpi.data?.kpi;

  const financeData = useMemo(
    () =>
      kpi
        ? [
            { label: t("stat_revenue"), value: kpi.revenue },
            { label: t("stat_expenses"), value: kpi.expenses },
            { label: t("stat_ebitda"), value: kpi.ebitda },
          ]
        : [
            { label: t("stat_revenue"), value: 0 },
            { label: t("stat_expenses"), value: 0 },
            { label: t("stat_ebitda"), value: 0 },
          ],
    [kpi, t],
  );

  const marginComposition = useMemo(
    () =>
      kpi
        ? [
            { label: t("stat_expenses"), value: kpi.expenses },
            { label: t("stat_ebitda"), value: kpi.ebitda },
          ]
        : [
            { label: t("stat_expenses"), value: 0 },
            { label: t("stat_ebitda"), value: 0 },
          ],
    [kpi, t],
  );

  const orgName =
    (orgList?.organizations ?? []).find((org) => String(org.id) === orgId)?.name ?? `#${orgId}`;

  const firstName = me?.user.firstName ?? "";
  const greeting = firstName ? t("greeting", { firstName }) : "";
  const stats = useDashboardStats();

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={t("title")}
        description={t("welcomeBack", { orgName })}
        actions={<DashboardActions />}
      />

      {greeting ? <p className="text-sm text-muted-foreground">{greeting}</p> : null}

      <section aria-label={t("keyMetrics")} className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {stats.map((stat) => (
          <StatCard
            key={stat.key}
            label={statLabel(stat.key)}
            value={stat.value}
            icon={stat.icon}
            trend={stat.trend}
            trendDirection={stat.trendDirection}
          />
        ))}
      </section>

      <div className="grid gap-4 lg:grid-cols-12">
        <Card className="lg:col-span-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="size-4" aria-hidden />
              {t("revenueSummaryTitle")}
            </CardTitle>
            <CardDescription>{t("revenueSummaryDesc")}</CardDescription>
          </CardHeader>
          <CardContent>
            <BarChart
              data={financeData}
              ariaLabel={t("revenueSummaryTitle")}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            />
          </CardContent>
        </Card>
        <Card className="lg:col-span-4 bg-muted/20">
          <CardHeader>
            <CardTitle>{t("stat_revenue")}</CardTitle>
            <CardDescription>{t("ledgerHealthDesc")}</CardDescription>
          </CardHeader>
          <CardContent>
            <DonutChart
              data={marginComposition}
              ariaLabel={t("revenueSummaryTitle")}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
              centerLabel={t("stat_revenue")}
            />
          </CardContent>
        </Card>
      </div>

      <ModulesGrid />

      <Separator />

      <p className="text-xs text-muted-foreground">{t("autoRefreshNote")}</p>
    </div>
  );
}
