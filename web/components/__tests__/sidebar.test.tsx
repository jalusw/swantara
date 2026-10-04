import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { LayoutDashboard } from "lucide-react";
import { afterEach, describe, expect, it } from "vitest";
import {
  Sidebar,
  SidebarContent,
  SidebarDivider,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/sidebar";
import { mediaQueryMock, renderWithProviders } from "@/lib/tests";

function TestSidebar() {
  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>Overview</SidebarGroupLabel>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton isActive label="Dashboard">
                  <LayoutDashboard />
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>
      <SidebarInset>
        <SidebarTrigger />
        <div>Main content</div>
      </SidebarInset>
    </SidebarProvider>
  );
}

describe("Sidebar", () => {
  it("renders an expanded sidebar and menu on desktop", () => {
    renderWithProviders(<TestSidebar />);

    const sidebar = document.querySelector("[data-slot='sidebar']");
    expect(sidebar).toHaveAttribute("data-state", "expanded");

    const button = screen.getByRole("button", { name: "Dashboard" });
    expect(button).toHaveAttribute("data-active", "true");
  });

  it("collapses and expands via the trigger on desktop", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TestSidebar />);

    const sidebar = document.querySelector("[data-slot='sidebar']");
    await user.click(screen.getByRole("button", { name: "Toggle sidebar" }));
    expect(sidebar).toHaveAttribute("data-state", "collapsed");

    await user.click(screen.getByRole("button", { name: "Toggle sidebar" }));
    expect(sidebar).toHaveAttribute("data-state", "expanded");
  });

  it("opens a mobile drawer via the trigger", async () => {
    mediaQueryMock.setMatches(true);
    const user = userEvent.setup();
    renderWithProviders(<TestSidebar />);

    await user.click(screen.getByRole("button", { name: "Toggle sidebar" }));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: "Dashboard" })).toBeInTheDocument();
  });
});

const STORAGE_KEY = "swantara:sidebar:width";

describe("Sidebar (extended)", () => {
  afterEach(() => {
    localStorage.removeItem(STORAGE_KEY);
  });

  function FullSidebar(collapsed = false) {
    return (
      <SidebarProvider defaultCollapsed={collapsed}>
        <Sidebar>
          <SidebarHeader>
            <span>Brand</span>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupLabel>Overview</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton isActive label="Dashboard">
                      <LayoutDashboard />
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
                <SidebarMenuAction aria-label="More actions" />
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarFooter>
            <span>Footer</span>
          </SidebarFooter>
        </Sidebar>
        <SidebarInset>
          <SidebarTrigger />
          <SidebarDivider />
          <div>Main content</div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  it("throws when sidebar parts are used outside the provider", () => {
    expect(() =>
      renderWithProviders(
        <SidebarMenuButton>
          <LayoutDashboard />
        </SidebarMenuButton>,
      ),
    ).toThrow("useSidebar must be used within <SidebarProvider>.");
  });

  it("renders every sidebar slot", () => {
    renderWithProviders(FullSidebar());
    for (const slot of [
      "sidebar-provider",
      "sidebar",
      "sidebar-header",
      "sidebar-content",
      "sidebar-footer",
      "sidebar-group",
      "sidebar-group-label",
      "sidebar-group-content",
      "sidebar-menu",
      "sidebar-menu-item",
      "sidebar-menu-button",
      "sidebar-menu-action",
      "sidebar-divider",
      "sidebar-inset",
    ]) {
      expect(document.querySelector(`[data-slot='${slot}']`)).toBeInTheDocument();
    }
    expect(screen.getByText("Brand")).toBeInTheDocument();
    expect(screen.getByText("Main content")).toBeInTheDocument();
  });

  it("restores the stored width and clamps it", () => {
    localStorage.setItem(STORAGE_KEY, "300");
    renderWithProviders(FullSidebar());
    expect(document.querySelector("[data-slot='sidebar']")).toHaveStyle({ width: "300px" });
  });

  it("falls back to the default when storage is invalid and clamps extremes", () => {
    localStorage.setItem(STORAGE_KEY, "not-a-number");
    const { unmount } = renderWithProviders(FullSidebar());
    expect(document.querySelector("[data-slot='sidebar']")).toHaveStyle({ width: "256px" });
    unmount();

    localStorage.setItem(STORAGE_KEY, "9999");
    renderWithProviders(FullSidebar());
    expect(document.querySelector("[data-slot='sidebar']")).toHaveStyle({ width: "480px" });
  });

  it("honors a clamped default width", () => {
    renderWithProviders(
      <SidebarProvider defaultWidth={10}>
        <Sidebar>
          <SidebarContent>Collapsed</SidebarContent>
        </Sidebar>
      </SidebarProvider>,
    );
    const sidebar = document.querySelector("[data-slot='sidebar']");
    expect(sidebar).toHaveAttribute("data-state", "collapsed");
    expect(sidebar).toHaveStyle({ width: "72px" });
  });

  it("persists width changes to storage", async () => {
    const user = userEvent.setup();
    renderWithProviders(FullSidebar());
    await user.click(screen.getByRole("button", { name: "Toggle sidebar" }));
    expect(localStorage.getItem(STORAGE_KEY)).toBe("72");
  });

  it("renders no resize handle", () => {
    renderWithProviders(FullSidebar());
    expect(
      document.querySelector("[data-slot='sidebar'] [aria-orientation='vertical']"),
    ).not.toBeInTheDocument();
  });

  it("hides subordinate parts when collapsed", () => {
    renderWithProviders(FullSidebar(true));
    expect(document.querySelector("[data-slot='sidebar-group-label']")).toHaveClass("sr-only");
    expect(document.querySelector("[data-slot='sidebar-menu-action']")).not.toBeInTheDocument();

    const button = screen.getByRole("button", { name: "Dashboard" });
    expect(button).toHaveAttribute("aria-label", "Dashboard");
  });

  it("clones asChild menu buttons", () => {
    renderWithProviders(
      <SidebarProvider>
        <Sidebar>
          <SidebarContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild isActive label="Dashboard">
                  <a href="/dashboard">Dashboard</a>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarContent>
        </Sidebar>
      </SidebarProvider>,
    );
    const dashboard = screen.getByRole("link", { name: "Dashboard" });
    expect(dashboard).toHaveAttribute("data-slot", "sidebar-menu-button");
    expect(dashboard).toHaveAttribute("href", "/dashboard");
  });

  it("activates a menu action on click", async () => {
    const user = userEvent.setup();
    let clicked = 0;
    renderWithProviders(
      <SidebarProvider>
        <Sidebar>
          <SidebarContent>
            <SidebarMenuAction
              aria-label="More actions"
              onClick={() => {
                clicked += 1;
              }}
            />
          </SidebarContent>
        </Sidebar>
      </SidebarProvider>,
    );
    await user.click(screen.getByRole("button", { name: "More actions" }));
    expect(clicked).toBe(1);
  });
});
