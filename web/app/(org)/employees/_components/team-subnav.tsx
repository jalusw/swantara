"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const TEAM_TABS = [
  { key: "employees", href: "/employees" },
  { key: "departments", href: "/employees/departments" },
  { key: "jobPositions", href: "/employees/job-positions" },
] as const;

export function TeamSubNav() {
  return <OrgSubNav label={"Team"} tabs={TEAM_TABS} />;
}
