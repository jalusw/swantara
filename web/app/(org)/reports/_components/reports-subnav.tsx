"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const REPORTS_TABS = [
  { key: "reports", href: "/reports" },
  { key: "trialBalance", href: "/reports/trial-balance" },
  { key: "agingReport", href: "/reports/aging" },
  { key: "profitAndLoss", href: "/reports/profit-and-loss" },
  { key: "balanceSheet", href: "/reports/balance-sheet" },
  { key: "cashFlowReport", href: "/reports/cash-flow" },
  {
    key: "inventoryValuation",
    href: "/reports/inventory-valuation",
  },
] as const;

export function ReportsSubNav() {
  return <OrgSubNav label={"Reports"} tabs={REPORTS_TABS} />;
}
