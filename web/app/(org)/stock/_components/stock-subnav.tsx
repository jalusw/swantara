"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const STOCK_TABS = [
  { key: "overview", href: "/stock" },
  { key: "warehouses", href: "/stock/warehouses" },
  { key: "shipments", href: "/stock/shipments" },
  { key: "counts", href: "/stock/counts" },
  { key: "transfers", href: "/stock/warehouse-transfers" },
  { key: "reorderRules", href: "/stock/reorder-rules" },
] as const;

export function StockSubNav() {
  return <OrgSubNav label={"Stock"} tabs={STOCK_TABS} />;
}
