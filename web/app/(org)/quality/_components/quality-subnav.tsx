"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const QUALITY_TABS = [
  {
    key: "quality",
    label: "Quality checks",
    href: "/quality",
  },
  { key: "qualityPoints", href: "/quality/points" },
  { key: "qualityAlerts", href: "/quality/alerts" },
] as const;

export function QualitySubNav() {
  return <OrgSubNav label={"Quality"} tabs={QUALITY_TABS} />;
}
