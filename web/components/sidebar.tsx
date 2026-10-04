"use client";

import { PanelLeft } from "lucide-react";
import {
  Children,
  type ComponentProps,
  cloneElement,
  createContext,
  isValidElement,
  type ReactElement,
  type ReactNode,
  type Ref,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useMediaQuery } from "usehooks-ts";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "./sheet";

const SIDEBAR_WIDTH_COLLAPSED = 72;
const SIDEBAR_WIDTH_EXPANDED = 256;
const SIDEBAR_WIDTH_MIN = 72;
const SIDEBAR_WIDTH_MAX = 480;
const SIDEBAR_STORAGE_KEY = "swantara:sidebar:width";

type SidebarState = "expanded" | "collapsed";

type SidebarContextValue = {
  state: SidebarState;
  width: number;
  setWidth: (w: number) => void;
  openMobile: boolean;
  setOpenMobile: (open: boolean) => void;
  isMobile: boolean;
  toggleSidebar: () => void;
};

const SidebarContext = createContext<SidebarContextValue | null>(null);

function useSidebar() {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar must be used within <SidebarProvider>.");
  }
  return context;
}

function clampWidth(w: number): number {
  return Math.min(SIDEBAR_WIDTH_MAX, Math.max(SIDEBAR_WIDTH_MIN, w));
}

function readStoredWidth(fallback: number): number {
  if (typeof window === "undefined") return fallback;
  try {
    const raw = localStorage.getItem(SIDEBAR_STORAGE_KEY);
    if (!raw) return fallback;
    const n = Number(raw);
    return Number.isFinite(n) ? clampWidth(n) : fallback;
  } catch {
    return fallback;
  }
}

function persistWidth(width: number) {
  try {
    localStorage.setItem(SIDEBAR_STORAGE_KEY, String(width));
  } catch {
    return;
  }
}

function SidebarProvider({
  defaultCollapsed = false,
  defaultWidth,
  className,
  children,
  ...props
}: ComponentProps<"div"> & {
  defaultCollapsed?: boolean;
  defaultWidth?: number;
}) {
  const isMobile = useMediaQuery("(max-width: 1023px)");
  const [width, setWidthState] = useState(() => {
    if (defaultWidth != null) return clampWidth(defaultWidth);
    const fallback = defaultCollapsed ? SIDEBAR_WIDTH_COLLAPSED : SIDEBAR_WIDTH_EXPANDED;
    return readStoredWidth(fallback);
  });
  const [openMobile, setOpenMobile] = useState(false);

  const collapsed = width <= SIDEBAR_WIDTH_COLLAPSED;

  useEffect(() => {
    if (!isMobile) {
      setOpenMobile(false);
    }
  }, [isMobile]);

  const setWidth = useCallback((w: number) => {
    setWidthState(clampWidth(w));
  }, []);

  useEffect(() => {
    persistWidth(width);
  }, [width]);

  const toggleSidebar = useCallback(() => {
    if (isMobile) {
      setOpenMobile((open) => !open);
    } else {
      setWidthState((prev) =>
        prev <= SIDEBAR_WIDTH_COLLAPSED ? SIDEBAR_WIDTH_EXPANDED : SIDEBAR_WIDTH_COLLAPSED,
      );
    }
  }, [isMobile]);

  const value = useMemo<SidebarContextValue>(
    () => ({
      state: collapsed ? "collapsed" : "expanded",
      width,
      setWidth,
      openMobile,
      setOpenMobile,
      isMobile,
      toggleSidebar,
    }),
    [collapsed, width, setWidth, openMobile, isMobile, toggleSidebar],
  );

  return (
    <SidebarContext.Provider value={value}>
      <div
        data-slot="sidebar-provider"
        className={cn("flex min-h-svh w-full", className)}
        {...props}
      >
        {children}
      </div>
    </SidebarContext.Provider>
  );
}

function Sidebar({ className, children, ...props }: ComponentProps<"aside">) {
  const { isMobile, state, width, openMobile, setOpenMobile } = useSidebar();

  if (isMobile) {
    return (
      <Sheet open={openMobile} onOpenChange={setOpenMobile}>
        <SheetContent side="left" showCloseButton={false} className="w-72 p-0">
          <SheetHeader className="sr-only">
            <SheetTitle>{"Sidebar"}</SheetTitle>
          </SheetHeader>
          <div
            data-slot="sidebar"
            data-state={state}
            className={cn(
              "flex h-full flex-col gap-0 bg-sidebar text-sidebar-foreground",
              className,
            )}
            {...(props as ComponentProps<"div">)}
          >
            {children}
          </div>
        </SheetContent>
      </Sheet>
    );
  }

  return (
    <aside
      data-slot="sidebar"
      data-state={state}
      className={cn(
        "sticky top-0 z-docked flex h-dvh shrink-0 flex-col gap-0 border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width] duration-200 ease-in-out",
        className,
      )}
      style={{ width }}
      {...props}
    >
      {children}
    </aside>
  );
}

function SidebarHeader({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      data-slot="sidebar-header"
      className={cn("flex shrink-0 flex-col gap-2 p-2", className)}
      {...props}
    />
  );
}

function SidebarContent({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      data-slot="sidebar-content"
      className={cn("flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-2 py-2", className)}
      {...props}
    />
  );
}

function SidebarFooter({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      data-slot="sidebar-footer"
      className={cn("flex shrink-0 flex-col gap-1 border-t border-sidebar-border p-2", className)}
      {...props}
    />
  );
}

function SidebarGroup({ className, ...props }: ComponentProps<"div">) {
  return (
    <div data-slot="sidebar-group" className={cn("flex flex-col gap-1", className)} {...props} />
  );
}

function SidebarGroupLabel({ className, ...props }: ComponentProps<"div">) {
  const { state } = useSidebar();
  return (
    <div
      data-slot="sidebar-group-label"
      className={cn(
        "px-2.5 py-1 text-[11px] font-medium tracking-wider text-muted-foreground uppercase",
        state === "collapsed" && "sr-only",
        className,
      )}
      {...props}
    />
  );
}

function SidebarGroupContent({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      data-slot="sidebar-group-content"
      className={cn("flex flex-col gap-1", className)}
      {...props}
    />
  );
}

function SidebarMenu({ className, ...props }: ComponentProps<"ul">) {
  return (
    <ul data-slot="sidebar-menu" className={cn("flex flex-col gap-1", className)} {...props} />
  );
}

function SidebarMenuItem({
  className,
  ref,
  ...props
}: ComponentProps<"li"> & { ref?: Ref<HTMLLIElement> }) {
  return (
    <li data-slot="sidebar-menu-item" ref={ref} className={cn("list-none", className)} {...props} />
  );
}

function SidebarMenuButton({
  asChild = false,
  isActive = false,
  className,
  label,
  children,
  ...props
}: ComponentProps<"button"> & {
  asChild?: boolean;
  isActive?: boolean;
  label?: ReactNode;
}) {
  const { state } = useSidebar();
  const collapsed = state === "collapsed";

  const classes = cn(
    "flex min-h-11 w-full items-center gap-2.5 rounded-md px-2.5 text-sm text-sidebar-foreground/70 outline-none transition-colors",
    "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
    "focus-visible:ring-3 focus-visible:ring-sidebar-ring/50",
    "data-[active=true]:bg-sidebar-primary/10 data-[active=true]:font-medium data-[active=true]:text-sidebar-primary",
    "[&>svg]:size-4 [&>svg]:shrink-0",
    collapsed && "size-11 justify-center px-0",
    className,
  );

  if (asChild && Children.count(children) === 1 && isValidElement(children)) {
    const child = Children.only(children) as ReactElement<{
      className?: string;
    }>;
    return cloneElement(child, {
      "data-slot": "sidebar-menu-button",
      "data-active": isActive,
      className: cn(classes, collapsed && "[&>span]:sr-only", child.props.className),
      ...(collapsed && typeof label === "string" ? { "aria-label": label } : {}),
      ...(props as Record<string, unknown>),
    } as React.HTMLAttributes<HTMLElement> & Record<string, unknown>);
  }

  return (
    <button
      type="button"
      data-slot="sidebar-menu-button"
      data-active={isActive}
      aria-label={collapsed && typeof label === "string" ? label : undefined}
      className={classes}
      {...props}
    >
      {children}
      {label ? <span className={cn("truncate", collapsed && "sr-only")}>{label}</span> : null}
    </button>
  );
}

function SidebarMenuAction({
  className,
  ...props
}: ComponentProps<"button"> & { showOnHover?: boolean }) {
  const { state } = useSidebar();
  if (state === "collapsed") return null;
  return (
    <button
      type="button"
      data-slot="sidebar-menu-action"
      className={cn(
        "flex size-11 shrink-0 items-center justify-center rounded-md text-sidebar-foreground/60 outline-none transition-colors",
        "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
        "focus-visible:ring-3 focus-visible:ring-sidebar-ring/50",
        "[[data-panel-open]>&]:bg-sidebar-accent [[data-panel-open]>&]:text-sidebar-foreground",
        className,
      )}
      {...props}
    />
  );
}

function SidebarTrigger({
  className,
  label,
  ...props
}: ComponentProps<typeof Button> & { label?: string }) {
  const { toggleSidebar } = useSidebar();
  return (
    <Button
      data-slot="sidebar-trigger"
      variant="ghost"
      size="icon-sm"
      aria-label={label ?? "Toggle sidebar"}
      onClick={toggleSidebar}
      className={className}
      {...props}
    >
      <PanelLeft />
    </Button>
  );
}

function SidebarDivider({ className, ...props }: ComponentProps<"hr">) {
  return (
    <hr
      data-slot="sidebar-divider"
      className={cn("mx-2 my-2 h-px shrink-0 border-0 bg-sidebar-border", className)}
      {...props}
    />
  );
}

function SidebarInset({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      data-slot="sidebar-inset"
      className={cn("flex min-h-svh min-w-0 flex-1 flex-col", className)}
      {...props}
    />
  );
}

export {
  SidebarProvider,
  Sidebar,
  SidebarHeader,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuAction,
  SidebarDivider,
  SidebarTrigger,
  SidebarInset,
  useSidebar,
};
