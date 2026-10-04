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
import { useTranslations } from "next-intl";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { cn } from "@/lib/utils";

export function ModulesGrid() {
  const { has } = usePermissions();
  const t = useTranslations("Dashboard");
  const moduleTitle = (key: string) =>
    (t as unknown as (k: string) => string)(`moduleTitle_${key}`);
  const moduleDesc = (key: string) => (t as unknown as (k: string) => string)(`moduleDesc_${key}`);

  const modules: Array<{
    key: string;
    href: string;
    icon: typeof Building2;
    permission?: string;
  }> = [
    {
      key: "crm",
      href: "/contacts",
      icon: ContactRound,
      permission: "contact.view",
    },
    {
      key: "sales",
      href: "/sale-orders/customers",
      icon: ShoppingCart,
      permission: "contact.view",
    },
    {
      key: "procurement",
      href: "/purchases/suppliers",
      icon: Truck,
      permission: "contact.view",
    },
    {
      key: "inventory",
      href: "/products",
      icon: Boxes,
      permission: "item.view",
    },
    {
      key: "manufacturing",
      href: "/products/recipes",
      icon: Factory,
      permission: "item.view",
    },
    {
      key: "accounting",
      href: "/accounting",
      icon: Calculator,
      permission: "journal_entry.view",
    },
    {
      key: "hr",
      href: "/employees",
      icon: UserRound,
      permission: "employee.view",
    },
    {
      key: "project",
      href: "/dashboard",
      icon: Briefcase,
      permission: "reporting.view",
    },
    {
      key: "subscription",
      href: "/subscriptions",
      icon: CreditCard,
      permission: "subscription.view",
    },
    {
      key: "quality",
      href: "/dashboard",
      icon: Layers,
      permission: "reporting.view",
    },
    {
      key: "returns",
      href: "/dashboard",
      icon: Receipt,
      permission: "reporting.view",
    },
    {
      key: "expenses",
      href: "/dashboard",
      icon: Wallet,
      permission: "reporting.view",
    },
    {
      key: "assets",
      href: "/dashboard",
      icon: Building2,
      permission: "reporting.view",
    },
    {
      key: "pos",
      href: "/dashboard",
      icon: ShoppingCart,
      permission: "reporting.view",
    },
    {
      key: "service",
      href: "/dashboard",
      icon: Settings,
      permission: "reporting.view",
    },
    {
      key: "reference",
      href: "/reference",
      icon: Layers,
      permission: "organization.view",
    },
  ];

  const visible = modules.filter((m) => !m.permission || has(m.permission));

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("exploreTitle")}</CardTitle>
        <CardDescription>{t("exploreDescription")}</CardDescription>
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
                  <span className="text-sm font-medium">{moduleTitle(mod.key)}</span>
                </span>
                <span className="line-clamp-2 text-sm leading-relaxed text-muted-foreground">
                  {moduleDesc(mod.key)}
                </span>
                <span className="mt-auto text-xs font-medium text-primary group-hover:underline">
                  {t("viewLink")} →
                </span>
              </a>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
