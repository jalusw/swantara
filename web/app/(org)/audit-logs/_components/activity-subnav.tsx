"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const ACTIVITY_TABS = [
  { key: "activity", label: "Audit logs", href: "/audit-logs" },
  {
    key: "integrationEvents",
    href: "/audit-logs/integration-events",
  },
] as const;

export function ActivitySubNav() {
  return <OrgSubNav label={"Activity"} tabs={ACTIVITY_TABS} />;
}
