"use client";

import { useTranslations } from "next-intl";
import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const EXPENSES_TABS = [
  { key: "expenses", href: "/expenses" },
  { key: "expenseCategories", href: "/expenses/categories" },
] as const;

export function ExpensesSubNav() {
  const tNav = useTranslations("Nav");
  return <OrgSubNav label={tNav("expenses")} tabs={EXPENSES_TABS} />;
}
