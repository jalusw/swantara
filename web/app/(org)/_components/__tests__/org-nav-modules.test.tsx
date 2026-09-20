import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { SidebarProvider } from "@/components/sidebar";
import { renderWithProviders, server } from "@/lib/tests";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { OrgNav } from "../org-nav";

function renderNav() {
  renderWithProviders(
    <OrgActiveProvider orgId={1}>
      <SidebarProvider>
        <OrgNav />
      </SidebarProvider>
    </OrgActiveProvider>,
  );
}

describe("OrgNav module gating", () => {
  it("hides sections of deactivated modules", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/organizations/:organizationId/modules", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: {
            modules: [
              { module_id: "crm", active: true },
              { module_id: "hr", active: false },
            ],
          },
        }),
      ),
    );
    renderNav();

    await user.click(await screen.findByRole("button", { name: "CRM" }));
    expect(await screen.findByRole("link", { name: "Sales" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "HR" })).not.toBeInTheDocument();
  });

  it("regression: groups former Other items into Finance and Settings with no Other menu", async () => {
    const user = userEvent.setup();
    renderNav();

    await user.click(await screen.findByRole("button", { name: "Finance" }));
    expect(await screen.findByRole("link", { name: "Reports" })).toHaveAttribute(
      "href",
      "/reports",
    );
    expect(screen.getByRole("link", { name: "Reference" })).toHaveAttribute("href", "/reference");
    expect(screen.queryByRole("button", { name: "Other" })).toBeNull();
    expect(screen.getByRole("link", { name: "Approval Requests" })).toHaveAttribute(
      "href",
      "/approval-requests",
    );
    expect(screen.getByRole("link", { name: "Activity" })).toHaveAttribute("href", "/audit-logs");
  });

  it("shows every permitted section when the modules request fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/v1/organizations/:organizationId/modules", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    renderNav();

    await user.click(await screen.findByRole("button", { name: "CRM" }));
    expect(await screen.findByRole("link", { name: "Sales" })).toBeInTheDocument();
    await user.click(await screen.findByRole("button", { name: "HR" }));
    expect(await screen.findByRole("link", { name: "Employees" })).toBeInTheDocument();
  });
});
