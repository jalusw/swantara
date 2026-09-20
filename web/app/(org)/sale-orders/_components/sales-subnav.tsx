"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const SALES_TABS = [
  { key: "sales", href: "/sale-orders" },
  { key: "customers", href: "/sale-orders/customers" },
  { key: "rmas", href: "/sale-orders/returns" },
] as const;

export function SalesSubNav() {
  return <OrgSubNav label={"Sales"} tabs={SALES_TABS} />;
}
