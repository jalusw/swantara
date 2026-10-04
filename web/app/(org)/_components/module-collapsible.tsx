"use client";

import {
  Briefcase,
  Calculator,
  ChevronDown,
  ClipboardCheck,
  CreditCard,
  Handshake,
  type LucideIcon,
  Package,
  Store,
  Truck,
  Users,
  Warehouse,
  Wrench,
} from "lucide-react";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/collapsible";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  useSidebar,
} from "@/components/sidebar";
import { cn } from "@/lib/utils";
import type { OrgNavItem, OrgNavModule } from "./nav-items";
import { SidebarTree } from "./sidebar-tree";

export const moduleKey: Record<OrgNavModule, string> = {
  crm: "CRM",
  products: "Produk",
  inventory: "Persediaan",
  procurement: "Pengadaan",
  hr: "HR",
  projects: "Proyek",
  quality: "Mutu",
  finance: "Keuangan",
  subscriptions: "Langganan",
  pos: "POS",
  service: "Layanan",
};

const moduleIcons: Record<OrgNavModule, LucideIcon> = {
  crm: Handshake,
  products: Package,
  inventory: Warehouse,
  procurement: Truck,
  hr: Users,
  projects: Briefcase,
  quality: ClipboardCheck,
  finance: Calculator,
  subscriptions: CreditCard,
  pos: Store,
  service: Wrench,
};

export type ModuleCollapsibleProps = {
  mod: OrgNavModule;
  items: OrgNavItem[];
};

export function ModuleCollapsible({ mod, items }: ModuleCollapsibleProps) {
  const pathname = usePathname();
  const { state, isMobile } = useSidebar();
  const collapsed = state === "collapsed" && !isMobile;
  const tNav = useTranslations("Nav");
  const tModules = useTranslations("NavModules");
  const ModuleIcon = moduleIcons[mod];
  const moduleLabel = (() => {
    try {
      return tModules(mod);
    } catch {
      return moduleKey[mod];
    }
  })();
  void tNav;
  const isModuleActive = items.some(
    (item) => pathname === item.href || pathname.startsWith(`${item.href}/`),
  );
  const [open, setOpen] = useState(isModuleActive);

  useEffect(() => {
    if (isModuleActive) setOpen(true);
  }, [isModuleActive]);

  if (collapsed) {
    return (
      <div data-slot="module-section">
        <SidebarGroup className="gap-1.5">
          <SidebarGroupContent>
            <SidebarTree items={items} />
          </SidebarGroupContent>
        </SidebarGroup>
      </div>
    );
  }

  return (
    <div data-slot="module-section">
      <Collapsible open={open} onOpenChange={setOpen}>
        <SidebarGroup className="gap-1.5">
          <CollapsibleTrigger
            render={<SidebarGroupLabel />}
            nativeButton={false}
            aria-label={moduleLabel}
            className={cn(
              "flex w-full cursor-pointer items-center gap-2 rounded-lg px-3 py-1.5 text-sm tracking-wide text-sidebar-foreground/60 uppercase",
              "justify-between",
              "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
              "focus-visible:ring-2 ring-offset-2 ring-offset-sidebar focus-visible:ring-sidebar-ring",
              "[data-panel-open]:bg-sidebar-accent/40 [data-panel-open]:text-sidebar-foreground",
            )}
          >
            <span className="flex min-w-0 items-center gap-2">
              <ModuleIcon className="size-5 shrink-0" />
              <span className="truncate">{moduleLabel}</span>
            </span>
            <ChevronDown
              className={cn(
                "size-4 shrink-0 transition-transform duration-200 [[data-panel-open]>&]:rotate-180",
              )}
            />
          </CollapsibleTrigger>
          <CollapsibleContent
            className={cn(
              "grid transition-all duration-200 ease-out",
              "data-[closed]:grid-rows-[0fr] data-[open]:grid-rows-[1fr]",
              "data-[closed]:[&>div]:overflow-hidden",
              "data-[closed]:[&>div]:opacity-0 data-[open]:[&>div]:opacity-100",
              "data-[closed]:[&>div]:transition-all data-[open]:[&>div]:transition-all",
              "data-[closed]:[&>div]:duration-200 data-[open]:[&>div]:duration-200",
              "data-[closed]:[&>div]:delay-0 data-[open]:[&>div]:delay-75",
            )}
          >
            <div className="min-h-0">
              <SidebarGroupContent className="ml-2 border-l border-sidebar-border/50 pl-2">
                <SidebarTree items={items} />
              </SidebarGroupContent>
            </div>
          </CollapsibleContent>
        </SidebarGroup>
      </Collapsible>
    </div>
  );
}
