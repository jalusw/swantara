"use client";

import { useTranslations } from "next-intl";
import { SidebarGroup, SidebarGroupContent, SidebarGroupLabel } from "@/components/sidebar";
import { useActiveModuleIds } from "@/lib/hooks/use-org-modules";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { ModuleCollapsible } from "./module-collapsible";
import {
  type OrgNavGroup,
  type OrgNavItem,
  type OrgNavModule,
  orgNavItems,
  orgNavModuleOrder,
} from "./nav-items";
import { SidebarTree } from "./sidebar-tree";

const groupOrder: OrgNavGroup[] = ["overview", "modules", "settings"];

export function OrgNav() {
  const { has } = usePermissions();
  const activeModules = useActiveModuleIds();
  const tNav = useTranslations("Nav");

  const GROUP_LABELS: Record<OrgNavGroup, string> = {
    overview: tNav("overview"),
    modules: tNav("modules"),
    settings: tNav("settings"),
  };

  const visibleItems = orgNavItems.filter(
    (item) =>
      (item.permission == null || has(item.permission)) &&
      (item.module == null || activeModules == null || activeModules.has(item.module)),
  );
  const overviewItems = visibleItems.filter((item) => item.group === "overview");
  const settingsItems = visibleItems.filter((item) => item.group === "settings");
  const moduleGroups = orgNavModuleOrder.filter((mod) =>
    visibleItems.some((item) => item.group === "modules" && item.module === mod),
  );

  function moduleItems(mod: OrgNavModule): OrgNavItem[] {
    return visibleItems.filter((item) => item.group === "modules" && item.module === mod);
  }

  return (
    <div className="flex flex-col gap-4">
      {groupOrder.map((group) => {
        if (group !== "modules") {
          const items = group === "overview" ? overviewItems : settingsItems;
          if (items.length === 0) return null;
          return (
            <SidebarGroup key={group}>
              <SidebarGroupLabel>{GROUP_LABELS[group]}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarTree items={items} />
              </SidebarGroupContent>
            </SidebarGroup>
          );
        }
        if (moduleGroups.length === 0) return null;
        return (
          <div key={group} className="contents">
            {moduleGroups.map((mod) => (
              <ModuleCollapsible key={mod} mod={mod} items={moduleItems(mod)} />
            ))}
          </div>
        );
      })}
    </div>
  );
}
