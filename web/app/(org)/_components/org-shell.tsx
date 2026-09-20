"use client";

import type { ReactNode } from "react";
import { AppShell } from "@/components/app-shell";

import { NotificationBell } from "@/components/notification-bell";
import { Separator } from "@/components/separator";
import { OrgBrand } from "./org-brand";
import { OrgBreadcrumbs } from "./org-breadcrumbs";
import { OrgNav } from "./org-nav";
import { OrgSearch } from "./org-search";
import { OrgUserMenu } from "./org-user-menu";

export function OrgShell({ children }: { children: ReactNode }) {
  return (
    <AppShell
      brand={<OrgBrand />}
      nav={<OrgNav />}
      search={<OrgSearch />}
      actions={
        <>
          <NotificationBell tooltip="Notifications" />
          <Separator orientation="vertical" className="h-6" />
          <OrgUserMenu />
        </>
      }
    >
      <div className="flex flex-col gap-6">
        <OrgBreadcrumbs />
        <div className="flex flex-col gap-6">{children}</div>
      </div>
    </AppShell>
  );
}
