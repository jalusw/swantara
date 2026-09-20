"use client";

import {
  ArrowRight,
  BarChart3,
  DollarSign,
  type LucideIcon,
  Package,
  Scale,
  TrendingUp,
  Wallet,
} from "lucide-react";
import Link from "next/link";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { usePermissions } from "@/lib/hooks/use-permissions";

type ReportCard = {
  key: string;
  title: string;
  description: string;
  href: string;
  icon: LucideIcon;
  permission?: string;
};

export function ReportsHubSection() {
  const { has } = usePermissions();

  const reports: ReportCard[] = [
    {
      key: "profitAndLoss",
      title: "Profit & Loss",
      description: "Income and expenses for the current period.",
      href: "/reports/profit-and-loss",
      icon: TrendingUp,
      permission: "journal_entry.view",
    },
    {
      key: "balanceSheet",
      title: "Balance Sheet",
      description: "Assets, liabilities, and equity snapshot.",
      href: "/reports/balance-sheet",
      icon: Scale,
      permission: "journal_entry.view",
    },
    {
      key: "cashFlow",
      title: "Cash Flow",
      description: "Cash movements across operating, investing, and financing.",
      href: "/reports/cash-flow",
      icon: Wallet,
      permission: "journal_entry.view",
    },
    {
      key: "trialBalance",
      title: "Trial Balance",
      description: "Debit and credit balances for all accounts.",
      href: "/reports/trial-balance",
      icon: BarChart3,
      permission: "journal_entry.view",
    },
    {
      key: "aging",
      title: "Accounts Aging",
      description: "Outstanding receivables by age bucket.",
      href: "/reports/aging",
      icon: DollarSign,
      permission: "journal_entry.view",
    },
    {
      key: "inventory",
      title: "Inventory Valuation",
      description: "Stock value by warehouse and SKU.",
      href: "/reports/inventory-valuation",
      icon: Package,
      permission: "item.view",
    },
  ];

  const visible = reports.filter((r) => !r.permission || has(r.permission));

  return (
    <section data-slot="reports-hub" className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {visible.map((report) => {
        const Icon = report.icon;
        return (
          <Link
            key={report.key}
            href={report.href}
            className="rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <Card className="group h-full transition-colors hover:bg-muted/50">
              <CardHeader>
                <div className="flex items-center justify-between gap-3">
                  <span className="grid size-10 place-items-center rounded-lg bg-primary/10 text-primary ring-1 ring-primary/15">
                    <Icon className="size-5" aria-hidden />
                  </span>
                  <ArrowRight
                    className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5"
                    aria-hidden
                  />
                </div>
                <CardTitle className="mt-1">{report.title}</CardTitle>
                <CardDescription>{report.description}</CardDescription>
              </CardHeader>
            </Card>
          </Link>
        );
      })}
    </section>
  );
}
