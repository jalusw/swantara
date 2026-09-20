"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const SERVICE_TABS = [
  { key: "service", href: "/service-orders" },
  { key: "equipments", href: "/service-orders/equipment" },
  { key: "serviceContracts", href: "/service-orders/contracts" },
  { key: "maintenancePlans", href: "/service-orders/plans" },
] as const;

export function ServiceSubNav() {
  return <OrgSubNav label={"Service"} tabs={SERVICE_TABS} />;
}
