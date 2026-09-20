"use client";

import {
  Boxes,
  Briefcase,
  Building2,
  Calculator,
  ContactRound,
  CreditCard,
  Factory,
  Layers,
  Receipt,
  Settings,
  ShoppingCart,
  Truck,
  UserRound,
  Wallet,
} from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { cn } from "@/lib/utils";

export function ModulesGrid() {
  const { has } = usePermissions();

  const modules: Array<{
    key: string;
    title: string;
    description: string;
    href: string;
    icon: typeof Building2;
    permission?: string;
  }> = [
    {
      key: "crm",
      title: "CRM",
      description: "Leads, opportunities and pipeline.",
      href: "/contacts",
      icon: ContactRound,
      permission: "contact.view",
    },
    {
      key: "sales",
      title: "Sales",
      description: "Quotations, orders, deliveries and invoicing.",
      href: "/sale-orders/customers",
      icon: ShoppingCart,
      permission: "contact.view",
    },
    {
      key: "procurement",
      title: "Procurement",
      description: "Requisitions, Quote Requests, purchase orders and receipts.",
      href: "/purchases/suppliers",
      icon: Truck,
      permission: "contact.view",
    },
    {
      key: "inventory",
      title: "Inventory",
      description: "Warehouses, stock moves and valuation.",
      href: "/products",
      icon: Boxes,
      permission: "item.view",
    },
    {
      key: "manufacturing",
      title: "Manufacturing",
      description: "BOMs, work orders and Planning.",
      href: "/products/recipes",
      icon: Factory,
      permission: "item.view",
    },
    {
      key: "accounting",
      title: "Accounting",
      description: "Chart of accounts, journals and posting.",
      href: "/accounting",
      icon: Calculator,
      permission: "journal_entry.view",
    },
    {
      key: "hr",
      title: "HR & Payroll",
      description: "Employees, leave, attendance and payroll.",
      href: "/employees",
      icon: UserRound,
      permission: "employee.view",
    },
    {
      key: "project",
      title: "Projects",
      description: "Projects, tasks and job costing.",
      href: "/dashboard",
      icon: Briefcase,
      permission: "reporting.view",
    },
    {
      key: "subscription",
      title: "Subscriptions",
      description: "Plans, subscribers and recurring billing.",
      href: "/subscriptions",
      icon: CreditCard,
      permission: "subscription.view",
    },
    {
      key: "quality",
      title: "Quality",
      description: "Points, checks and alerts.",
      href: "/dashboard",
      icon: Layers,
      permission: "reporting.view",
    },
    {
      key: "returns",
      title: "Returns",
      description: "RMAs and credit notes.",
      href: "/dashboard",
      icon: Receipt,
      permission: "reporting.view",
    },
    {
      key: "expenses",
      title: "Expenses",
      description: "Reports, approvals and reimbursements.",
      href: "/dashboard",
      icon: Wallet,
      permission: "reporting.view",
    },
    {
      key: "assets",
      title: "Fixed Assets",
      description: "Registers, depreciation and disposal.",
      href: "/dashboard",
      icon: Building2,
      permission: "reporting.view",
    },
    {
      key: "pos",
      title: "Point of Sale",
      description: "Configs, sessions and store orders.",
      href: "/dashboard",
      icon: ShoppingCart,
      permission: "reporting.view",
    },
    {
      key: "service",
      title: "Service",
      description: "Contracts, orders and warranties.",
      href: "/dashboard",
      icon: Settings,
      permission: "reporting.view",
    },
    {
      key: "reference",
      title: "Reference data",
      description: "Currencies, UoMs and dimension accounts.",
      href: "/reference",
      icon: Layers,
      permission: "organization.view",
    },
  ];

  const visible = modules.filter((m) => !m.permission || has(m.permission));

  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Explore business flows"}</CardTitle>
        <CardDescription>
          {"Every service your organization runs — one tap to its module."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid auto-rows-[minmax(140px,auto)] gap-4 grid-cols-12">
          {visible.map((mod, index) => {
            const Icon = mod.icon;
            const span =
              index === 0
                ? "col-span-12 md:col-span-6 lg:col-span-8"
                : index < 3
                  ? "col-span-12 md:col-span-6 lg:col-span-4"
                  : index < 9
                    ? "col-span-12 md:col-span-6 lg:col-span-4"
                    : "col-span-12 md:col-span-6 lg:col-span-3";
            return (
              <a
                key={mod.key}
                href={mod.href}
                className={cn(
                  "group flex flex-col gap-2 rounded-xl border border-border p-4 transition-colors hover:bg-muted/50",
                  span,
                  index === 0 && "bg-primary/[0.04] border-primary/15",
                )}
              >
                <span className="flex items-center gap-2">
                  <span className="grid size-8 place-items-center rounded-lg bg-primary/10 text-primary ring-1 ring-primary/15">
                    <Icon className="size-4" aria-hidden />
                  </span>
                  <span className="text-sm font-medium">{mod.title}</span>
                </span>
                <span className="line-clamp-2 text-sm leading-relaxed text-muted-foreground">
                  {mod.description}
                </span>
                <span className="mt-auto text-xs font-medium text-primary group-hover:underline">
                  {"View"} →
                </span>
              </a>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
