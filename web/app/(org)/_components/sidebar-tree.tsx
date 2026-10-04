"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "@/components/sidebar";
import type { OrgNavItem } from "./nav-items";
import { navLabel } from "./nav-items";

export type SidebarTreeProps = {
  items: OrgNavItem[];
};

export function SidebarTree({ items }: SidebarTreeProps) {
  const pathname = usePathname();
  const tNav = useTranslations("Nav");

  function getLabel(key: string): string {
    try {
      // next-intl will throw if key is missing; fall back to legacy Indonesian label
      return (tNav as unknown as (k: string) => string)(key);
    } catch {
      return navLabel(key);
    }
  }

  return (
    <SidebarMenu>
      {items.map((item) => {
        const Icon = item.icon;
        const href = item.href;
        const label = getLabel(item.key);
        const isActive = pathname === href;
        return (
          <SidebarMenuItem key={item.key}>
            <SidebarMenuButton asChild isActive={isActive} label={label}>
              <Link href={href}>
                <Icon />
                <span>{label}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        );
      })}
    </SidebarMenu>
  );
}
