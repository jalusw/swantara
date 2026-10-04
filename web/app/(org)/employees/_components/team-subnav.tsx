"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const TEAM_TABS = [
  { key: "employees", href: "/employees" },
  { key: "departments", href: "/employees/departments" },
  { key: "jobPositions", href: "/employees/job-positions" },
] as const;

export function TeamSubNav() {
  const t = useTranslations("Employees");
  return <OrgSubNav label={t("teamLabel")} tabs={TEAM_TABS} />;
}
