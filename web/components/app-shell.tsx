"use client";

import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "./sidebar";

export type AppShellLabels = {
  toggleSidebar?: string;
};

export type AppShellProps = {
  /** Content rendered at the top of the sidebar (logo, org switcher). */
  brand?: ReactNode;
  /** Sidebar navigation content (SidebarGroup + SidebarMenu). */
  nav: ReactNode;
  /** Content rendered at the bottom of the sidebar (user, collapse). */
  footer?: ReactNode;
  /** Search control rendered in the header. */
  search?: ReactNode;
  /** Actions rendered at the end of the header (theme, user menu). */
  actions?: ReactNode;
  /** Accessible labels used by the shell's sidebar controls. */
  labels?: AppShellLabels;
  /** Start the desktop sidebar collapsed to icons. */
  defaultCollapsed?: boolean;
  className?: string;
  children: ReactNode;
};

export function AppShell({
  brand,
  nav,
  footer,
  search,
  actions,
  labels,
  defaultCollapsed = false,
  className,
  children,
}: AppShellProps) {
  return (
    <SidebarProvider data-slot="app-shell" defaultCollapsed={defaultCollapsed}>
      <Sidebar>
        {brand ? <SidebarHeader>{brand}</SidebarHeader> : null}
        <SidebarContent>{nav}</SidebarContent>
        {footer ? <SidebarFooter>{footer}</SidebarFooter> : null}
      </Sidebar>
      <SidebarInset className="bg-muted/20">
        <header className="sticky top-0 z-header flex h-[64px] shrink-0 items-center gap-3 border-b border-border bg-background px-4 sm:h-[64px] sm:px-6">
          <SidebarTrigger label={labels?.toggleSidebar} />
          {search ? (
            <div className="flex min-w-0 flex-1 justify-start">{search}</div>
          ) : (
            <div className="flex-1" />
          )}
          {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
        </header>
        {/* biome-ignore lint/correctness/useUniqueElementIds: skip navigation target must be static */}
        <main
          id="main"
          className={cn(
            "mx-auto flex w-full max-w-screen-2xl flex-1 flex-col p-4 sm:p-6 lg:p-8",
            className,
          )}
        >
          <div className="flex flex-1 flex-col gap-6">{children}</div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
