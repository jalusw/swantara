import { screen } from "@testing-library/react";
import { Layers, LayoutDashboard, Package } from "lucide-react";
import { describe, expect, it } from "vitest";
import { SidebarProvider } from "@/components/sidebar";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import type { OrgNavItem } from "../nav-items";
import { SidebarTree } from "../sidebar-tree";

const ITEMS: OrgNavItem[] = [
  {
    key: "dashboard",
    group: "overview",
    href: "/dashboard",
    icon: LayoutDashboard,
  },
  {
    key: "products",
    group: "modules",
    module: "products",
    href: "/products",
    icon: Package,
  },
  {
    key: "recipes",
    group: "modules",
    module: "products",
    href: "/products/recipes",
    icon: Layers,
  },
];

function renderTree(props: Partial<React.ComponentProps<typeof SidebarTree>> = {}) {
  return renderWithProviders(
    <SidebarProvider>
      <SidebarTree items={ITEMS} {...props} />
    </SidebarProvider>,
  );
}

describe("SidebarTree branches2", () => {
  it("marks a former child link as active", () => {
    navigationMock.setPathname("/products/recipes");
    renderTree();

    expect(screen.getByRole("link", { name: "Bills of materials" })).toHaveAttribute(
      "data-active",
      "true",
    );
    expect(screen.getByRole("link", { name: "Dashboard" })).toHaveAttribute("data-active", "false");
  });

  it("renders nothing extra for an empty item list", () => {
    renderTree({ items: [] });

    expect(screen.queryByRole("link")).toBeNull();
  });

  it("renders former branch children as flat siblings", () => {
    renderTree();

    expect(screen.getByRole("link", { name: "Bills of materials" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Dashboard" })).toBeInTheDocument();
  });

  it("regression: renders no nested submenu elements", () => {
    renderTree();

    expect(document.querySelector("[data-slot='sidebar-menu-sub']")).toBeNull();
    expect(document.querySelector("[data-slot='sidebar-menu-sub-item']")).toBeNull();
    expect(document.querySelector("[data-slot='sidebar-menu-sub-button']")).toBeNull();
  });
});
