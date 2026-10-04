import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import { OrgBreadcrumbs } from "../org-breadcrumbs";

describe("OrgBreadcrumbs", () => {
  it("renders the trail from the pathname", () => {
    navigationMock.setPathname("/settings/members");
    renderWithProviders(<OrgBreadcrumbs />);

    expect(screen.getByRole("link", { name: "Beranda" })).toHaveAttribute("href", "/dashboard");
    expect(screen.getByRole("link", { name: "Umum" })).toHaveAttribute("href", "/settings");
    expect(screen.getByText("Anggota")).toHaveAttribute("aria-current", "page");
  });

  it("marks the dashboard as the current crumb on the dashboard", () => {
    navigationMock.setPathname("/dashboard");
    renderWithProviders(<OrgBreadcrumbs />);

    expect(screen.queryByRole("link", { name: "Beranda" })).toBeNull();
    expect(screen.getByText("Dasbor")).toHaveAttribute("aria-current", "page");
  });

  it("renders nothing on the root path", () => {
    navigationMock.setPathname("/");
    const { container } = renderWithProviders(<OrgBreadcrumbs />);

    expect(container).toBeEmptyDOMElement();
  });
});
