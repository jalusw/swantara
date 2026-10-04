"use client";

import { Palmtree, UserCheck, UserRound, UsersRound } from "lucide-react";
import { useTranslations } from "next-intl";
import { MetricGrid } from "@/components/metric-grid";
import { StatCard } from "@/components/stat-card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Employee, LeaveRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

function isCurrentlyOnLeave(request: LeaveRequest, now: Date): boolean {
  if (request.state !== "approved") return false;
  const from = new Date(request.dateFrom);
  const to = new Date(request.dateTo);
  return from <= now && to >= now;
}

function isHiredThisMonth(employee: Employee, now: Date): boolean {
  if (!employee.hireDate) return false;
  const hireDate = new Date(employee.hireDate);
  return hireDate.getMonth() === now.getMonth() && hireDate.getFullYear() === now.getFullYear();
}

export function EmployeesStats() {
  const t = useTranslations("Employees");
  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );

  const leaveRequestsQuery = useOrgListQuery<
    { leaveRequests: LeaveRequest[] },
    Record<string, never>
  >("leaveRequests", (organizationId) => getSwantaraService().leaveRequests.list(organizationId));

  const employees = employeesQuery.data?.employees ?? [];
  const leaveRequests = leaveRequestsQuery.data?.leaveRequests ?? [];
  const now = new Date();

  const activeCount = employees.filter((e) => e.active).length;
  const onLeaveCount = leaveRequests.filter((lr) => isCurrentlyOnLeave(lr, now)).length;
  const newThisMonthCount = employees.filter((e) => isHiredThisMonth(e, now)).length;

  return (
    <MetricGrid>
      <StatCard label={t("statTotal")} value={formatNumber(employees.length)} icon={UsersRound} />
      <StatCard label={t("statActive")} value={formatNumber(activeCount)} icon={UserCheck} />
      <StatCard label={t("statOnLeave")} value={formatNumber(onLeaveCount)} icon={Palmtree} />
      <StatCard
        label={t("statNewThisMonth")}
        value={formatNumber(newThisMonthCount)}
        icon={UserRound}
      />
    </MetricGrid>
  );
}
