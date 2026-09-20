"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "@/components/sidebar";
import type { OrgNavItem } from "./nav-items";
import { navLabel } from "./nav-items";

export type SidebarTreeProps = {
  items: OrgNavItem[];
};

export function SidebarTree({ items }: SidebarTreeProps) {
  const pathname = usePathname();

  return (
    <SidebarMenu>
      {items.map((item) => {
        const Icon = item.icon;
        const href = item.href;
        const label = navLabel(item.key);
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
