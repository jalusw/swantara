"use client";

import { usePathname } from "next/navigation";
import { Breadcrumb, type BreadcrumbItem } from "@/components/breadcrumb";
import { humanizeKey } from "@/lib/utils/case";
import { navLabel, orgNavItems } from "./nav-items";

export function OrgBreadcrumbs() {
  const pathname = usePathname();

  const trail = pathname.split("/").filter(Boolean);
  if (trail.length === 0) {
    return null;
  }
  const items: BreadcrumbItem[] = [
    {
      label: "Home",
      href: trail[0] === "dashboard" ? undefined : "/dashboard",
    },
  ];

  const crmTabKeys = new Set(["pipeline", "leads", "opportunities", "activities"]);
  const settingsSubPageKeys = new Set(["profile"]);
  for (let i = 1; i <= trail.length; i++) {
    const prefix = `/${trail.slice(0, i).join("/")}`;
    const match = orgNavItems.find((item) => item.href === prefix);
    const isLast = i === trail.length;
    let label: string;
    if (match) {
      label = navLabel(match.key);
    } else if (trail[0] === "crm" && crmTabKeys.has(trail[i - 1]!)) {
      try {
        label = humanizeKey(String(trail[i - 1]!));
      } catch {
        label = trail[i - 1]!;
      }
    } else if (trail[0] === "settings" && settingsSubPageKeys.has(trail[i - 1]!)) {
      try {
        label = humanizeKey(String(trail[i - 1]!));
      } catch {
        label = trail[i - 1]!;
      }
    } else {
      label = prefix;
    }
    items.push({
      label,
      href: isLast ? undefined : prefix,
    });
  }

  return <Breadcrumb items={items} />;
}
