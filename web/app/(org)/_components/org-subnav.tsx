"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { navLabel } from "./nav-items";

export type OrgSubNavTab = {
  key: string;
  label?: string;
  href: string;
};

export function OrgSubNav({ label, tabs }: { label: string; tabs: readonly OrgSubNavTab[] }) {
  const pathname = usePathname();
  const hrefs = tabs.map((tab) => tab.href);

  function isActive(href: string): boolean {
    if (pathname === href) return true;
    if (!pathname.startsWith(`${href}/`)) return false;
    return !hrefs.some((other) => other !== href && pathname.startsWith(other));
  }

  return (
    <nav aria-label={label}>
      <div className="inline-flex w-full items-center gap-6 overflow-x-auto border-b border-border">
        {tabs.map((tab) => {
          const href = tab.href;
          const active = isActive(href);
          return (
            <Link
              key={tab.key}
              href={href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "relative inline-flex min-h-11 shrink-0 items-center border-b-2 px-1 text-sm whitespace-nowrap transition-colors",
                active
                  ? "border-primary text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground",
              )}
            >
              {tab.label ?? navLabel(tab.key)}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
