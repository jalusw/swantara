import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProjectDetail } from "../project-detail-section";

function projectFixture(overrides = {}) {
  return {
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
    ...overrides,
  };
}

const summaryFixture = {
  planned_hours: 100,
  effective_hours: 40,
  billable_amount: 4000,
  billed_amount: 1000,
  unbilled_amount: 3000,
  unbilled_hours: 30,
  cost: 2000,
  cost_amount: 2000,
  revenue: 4000,
  margin_amount: 2000,
};

function useProjectHandlers(project: Record<string, unknown>, summary: unknown) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { project } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:id/summary", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { summary } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/tasks", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { tasks: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/milestones", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { milestones: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/timesheets", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { timesheets: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          contacts: [{ id: 10, organization_id: 1, name: "Acme Corp", display_name: "Acme Corp" }],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
  );
}

beforeEach(() => {});

describe("ProjectDetail branches4", () => {
  it("hides state actions for closed projects", async () => {
    useProjectHandlers(projectFixture({ state: "closed" }), summaryFixture);
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    expect(screen.queryByRole("button", { name: "Batal" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Action Buka" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Action Tutup" })).toBeNull();
  });

  it("renders an empty billing panel when the summary is missing", async () => {
    useProjectHandlers(projectFixture({ billing_type: "time_material" }), null);
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    expect(screen.queryByText("Jam rencana")).toBeNull();
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(screen.queryByText("Jam belum ditagih")).toBeNull();
  });

  it("falls back to a unit rate when the billable rate is zero", async () => {
    useProjectHandlers(
      projectFixture({ billing_type: "time_material", billable_rate: 0 }),
      summaryFixture,
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(await screen.findByText("Jam belum ditagih")).toBeInTheDocument();
  });
});
