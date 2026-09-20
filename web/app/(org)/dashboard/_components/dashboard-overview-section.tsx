"use client";

import { Banknote, BarChart3, Truck, UserRound, Users } from "lucide-react";
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
import { humanizeKey } from "@/lib/utils/case";
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
        ? `${(finance.netMarginPct * 100).toFixed(0)}% from last month`
        : "—",
      icon: Banknote,
      permission: "journal_entry.view",
    },
    {
      key: "customers",
      value: formatNumber(contactsCount),
      trend: `${"—"} from last month`,
      icon: Users,
      permission: "contact.view",
    },
    {
      key: "suppliers",
      value: formatNumber(inventory?.productCount ?? contactsCount),
      trend: `${"—"} from last month`,
      icon: Truck,
      permission: "contact.view",
    },
    {
      key: "employees",
      value: formatNumber(employeesCount || procurement?.purchaseCount || 0),
      trend: `${"—"} from last month`,
      icon: UserRound,
      permission: "employee.view",
    },
  ];

  return stats.filter((stat) => has(stat.permission));
}

export function DashboardOverview({ orgId }: { orgId: string }) {
  const { data: me } = useMeQuery();
  const { data: orgList } = useMeOrganizationsQuery();
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

  const marginComposition = useMemo(
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

  const orgName =
    (orgList?.organizations ?? []).find((org) => String(org.id) === orgId)?.name ?? `#${orgId}`;

  const firstName = me?.user.firstName ?? "";
  const greeting = firstName ? `Good to see you, ${firstName}!` : "";
  const stats = useDashboardStats();

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Dashboard"}
        description={`Welcome back to ${orgName}. Here's what's happening across your business today.`}
        actions={<DashboardActions />}
      />

      {greeting ? <p className="text-sm text-muted-foreground">{greeting}</p> : null}

      <section aria-label={"Key metrics"} className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {stats.map((stat) => (
          <StatCard
            key={stat.key}
            label={humanizeKey(String(stat.key))}
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
              {"Revenue overview"}
            </CardTitle>
            <CardDescription>{"Monthly revenue for the current tax year."}</CardDescription>
          </CardHeader>
          <CardContent>
            <BarChart
              data={financeData}
              ariaLabel={"Revenue overview"}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
            />
          </CardContent>
        </Card>
        <Card className="lg:col-span-4 bg-muted/20">
          <CardHeader>
            <CardTitle>{"Revenue"}</CardTitle>
            <CardDescription>{"General ledger health and margins."}</CardDescription>
          </CardHeader>
          <CardContent>
            <DonutChart
              data={marginComposition}
              ariaLabel={"Revenue overview"}
              valueFormatter={(value) => formatMoney(value, { currency: DEFAULT_CURRENCY })}
              centerLabel={"Revenue"}
            />
          </CardContent>
        </Card>
      </div>

      <ModulesGrid />

      <Separator />

      <p className="text-xs text-muted-foreground">
        {"Data refreshes automatically. Last synced just now."}
      </p>
    </div>
  );
}
