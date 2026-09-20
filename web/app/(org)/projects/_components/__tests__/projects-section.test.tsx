import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProjectsSection } from "../projects-section";

function useLocalProjects() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/projects", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          projects: [
            {
              id: 1,
              organization_id: 1,
              name: "Website Redesign",
              contact_id: 10,
              manager_id: null,
              dimension_id: null,
              sale_order_id: null,
              billing_type: "fixed",
              billable_rate: 100,
              date_start: null,
              date_end: null,
              state: "open",
            },
            {
              id: 2,
              organization_id: 1,
              name: "Mobile App",
              contact_id: 11,
              manager_id: null,
              dimension_id: null,
              sale_order_id: null,
              billing_type: "time_material",
              billable_rate: 80,
              date_start: null,
              date_end: null,
              state: "draft",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          contacts: [
            { id: 10, organization_id: 1, name: "Acme Corp", display_name: "Acme Corp" },
            { id: 11, organization_id: 1, name: "Globex", display_name: "Globex" },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/dimensions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/sale-orders", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { orders: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalProjects();
});

describe("ProjectsSection", () => {
  it("renders seeded projects", async () => {
    renderWithProviders(<ProjectsSection orgId="1" />);

    expect(await screen.findByText("Website Redesign")).toBeInTheDocument();
    expect(screen.getByText("Mobile App")).toBeInTheDocument();
  });

  it("filters projects by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProjectsSection orgId="1" />);

    await screen.findByText("Website Redesign");
    await user.type(screen.getByPlaceholderText("Search projects…"), "Mobile");

    expect(await screen.findByText("Mobile App")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProjectsSection orgId="1" />);

    await screen.findByText("Website Redesign");
    await user.click(screen.getByRole("button", { name: "New project" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
