"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const COMMISSIONS_TABS = [
  { key: "commissions", href: "/commission-plans" },
  { key: "commissionEntries", href: "/commission-plans/entries" },
] as const;

export function CommissionsSubNav() {
  return <OrgSubNav label={"Commissions"} tabs={COMMISSIONS_TABS} />;
}
