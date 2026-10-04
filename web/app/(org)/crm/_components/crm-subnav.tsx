"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

const CRM_TABS = [
  { key: "pipeline", href: "/crm" },
  { key: "leads", href: "/crm/leads" },
  {
    key: "opportunities",
    href: "/crm/opportunities",
  },
  {
    key: "activities",
    href: "/crm/activities",
  },
] as const;

export function CrmSubNav() {
  const tNav = useTranslations("Nav");
  const pathname = usePathname();

  function tabLabel(key: string): string {
    try {
      return (tNav as unknown as (k: string) => string)(key);
    } catch {
      return humanizeKey(String(key));
    }
  }

  return (
    <nav aria-label={tabLabel("pipeline")}>
      <div className="inline-flex w-full items-center gap-6 border-b border-border">
        {CRM_TABS.map((tab) => {
          const href = tab.href;
          const isActive =
            tab.key === "pipeline"
              ? pathname === href
              : pathname === href || pathname.startsWith(`${href}/`);
          return (
            <Link
              key={tab.key}
              href={href}
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "relative inline-flex min-h-11 items-center border-b-2 px-1 text-sm whitespace-nowrap transition-colors",
                isActive
                  ? "border-primary text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {tabLabel(String(tab.key))}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
