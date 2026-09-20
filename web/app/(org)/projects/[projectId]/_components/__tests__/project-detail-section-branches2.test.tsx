import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProjectDetail } from "../project-detail-section";

const project = {
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
};

const summary = {
  planned_hours: 100,
  effective_hours: 40,
  billable_amount: 4000,
  billed_amount: 1000,
  cost: 2000,
  revenue: 4000,
};

function seedProject(overrides = {}) {
  const {
    project: projectOverride = {},
    tasks = [],
    milestones = [],
    timesheets = [],
    summary: summaryOverride = summary,
  } = overrides as {
    project?: Record<string, unknown>;
    tasks?: unknown[];
    milestones?: unknown[];
    timesheets?: unknown[];
    summary?: unknown;
  };
  server.use(
    http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { project: { ...project, ...projectOverride } },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:id/summary", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { summary: summaryOverride } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/tasks", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { tasks } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/milestones", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { milestones } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/timesheets", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { timesheets } }),
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

describe("ProjectDetail branches2", () => {
  it("renders the not-found branch when the project is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { project: null } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/projects/:id/summary", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { summary: null } }),
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
        HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
      ),
    );
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    expect(await screen.findByText("Project not found.")).toBeInTheDocument();
  });

  it("renders empty tasks, milestones and timesheets branches", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tasks" }));
    expect(await screen.findByText("No tasks yet.")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Milestones" }));
    expect(await screen.findByText("No milestones yet.")).toBeInTheDocument();
  });

  it("renders task deadline and milestone reached branches", async () => {
    seedProject({
      tasks: [
        { id: 1, project_id: 1, name: "Design", state: "open", deadline: "2026-03-01" },
        { id: 2, project_id: 1, name: "Build", state: "done", deadline: null },
      ],
      milestones: [
        { id: 1, project_id: 1, name: "Kickoff", reached: true, deadline: "2026-02-01" },
        { id: 2, project_id: 1, name: "Launch", reached: false, deadline: null },
      ],
    });
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tasks" }));
    expect(await screen.findByText("Design")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Milestones" }));
    expect(await screen.findByText("Kickoff")).toBeInTheDocument();
  });

  it("falls back to dash for unknown manager and missing summary", async () => {
    seedProject({ project: { manager_id: 99 }, summary: null });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });
});
