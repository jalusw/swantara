"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const EXPENSES_TABS = [
  { key: "expenses", href: "/expenses" },
  { key: "expenseCategories", href: "/expenses/categories" },
] as const;

export function ExpensesSubNav() {
  return <OrgSubNav label={"Expenses"} tabs={EXPENSES_TABS} />;
}
