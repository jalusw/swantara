"use client";

import { OrgSubNav } from "@/app/(org)/_components/org-subnav";

const ACCOUNTING_TABS = [
  { key: "overview", href: "/accounting" },
  { key: "accounts", href: "/accounting/accounts" },
  { key: "journals", href: "/accounting/journals" },
  { key: "taxes", href: "/accounting/taxes" },
  { key: "taxYears", href: "/accounting/tax-years" },
  { key: "invoices", href: "/accounting/invoices" },
  { key: "payments", href: "/accounting/payments" },
  { key: "bankStatements", href: "/accounting/bank-statements" },
  { key: "journalEntries", href: "/accounting/journal-entries" },
] as const;

export function AccountingSubNav() {
  return <OrgSubNav label={"Accounting"} tabs={ACCOUNTING_TABS} />;
}
