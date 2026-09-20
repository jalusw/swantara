"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const PURCHASES_TABS = [
  { key: "purchases", href: "/purchases" },
  { key: "purchaseRequests", href: "/purchases/requisitions" },
  { key: "supplierQuoteRequests", href: "/purchases/quoteRequests" },
  { key: "inboundCosts", href: "/purchases/inbound-costs" },
  { key: "suppliers", href: "/purchases/suppliers" },
] as const;

export function PurchasesSubNav() {
  return <OrgSubNav label={"Purchases"} tabs={PURCHASES_TABS} />;
}
