"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const PAYROLL_TABS = [
  { key: "payrollRuns", href: "/payroll-runs" },
  { key: "salaryRules", href: "/payroll-runs/salary-rules" },
  { key: "payslips", href: "/payroll-runs/payslips" },
] as const;

export function PayrollSubNav() {
  const t = useTranslations("Payroll");
  return <OrgSubNav label={t("subnavLabel")} tabs={PAYROLL_TABS} />;
}
