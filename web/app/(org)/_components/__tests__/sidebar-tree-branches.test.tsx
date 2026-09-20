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

describe("SidebarTree", () => {
  it("renders every link flat without nested submenus", () => {
    renderTree();

    expect(screen.getByRole("link", { name: "Dashboard" })).toHaveAttribute("href", "/dashboard");
    expect(screen.getByRole("link", { name: "Products" })).toHaveAttribute("href", "/products");
    expect(screen.getByRole("link", { name: "Bills of materials" })).toHaveAttribute(
      "href",
      "/products/recipes",
    );
  });

  it("marks the link matching the current pathname as active", () => {
    navigationMock.setPathname("/dashboard");
    renderTree();

    expect(screen.getByRole("link", { name: "Dashboard" })).toHaveAttribute("data-active", "true");
    expect(screen.getByRole("link", { name: "Products" })).toHaveAttribute("data-active", "false");
  });

  it("renders items in definition order", () => {
    renderTree();

    const links = screen.getAllByRole("link").map((link) => link.textContent);
    expect(links.indexOf("Dashboard")).toBeLessThan(links.indexOf("Products"));
  });
});
