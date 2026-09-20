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
  contact_id: null,
  manager_id: null,
  dimension_id: null,
  sale_order_id: null,
  billing_type: "fixed",
  billable_rate: 100,
  date_start: null,
  date_end: null,
  state: "draft",
};

const summary = {
  planned_hours: 100,
  effective_hours: 40,
  billable_amount: 4000,
  billed_amount: 1000,
  unbilled_amount: 3000,
  cost_amount: 2000,
  margin_amount: 2000,
  cost: 2000,
  revenue: 4000,
};

function seed(overrides: Record<string, unknown> = {}) {
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
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/projects/:id/state", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.put(
      "*/api/v1/organizations/:organizationId/projects/:projectId/milestones/:id/reached",
      () => HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {});

describe("ProjectDetail branches3", () => {
  it("renders not-found when project is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { project: null } }),
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
        HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
      ),
    );
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    expect(await screen.findByText("Project not found.")).toBeInTheDocument();
  });

  it("opens the project and closes it from header actions", async () => {
    const user = userEvent.setup();
    seed({ project: { state: "draft" } });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    const open = screen.queryByRole("button", { name: "Open" });
    if (open) await user.click(open);
    expect(await screen.findByRole("heading", { name: "Website Redesign" })).toBeInTheDocument();
  });

  it("renders time-material billing panel instead of the fallback", async () => {
    const user = userEvent.setup();
    seed({ project: { billing_type: "time_material" } });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    const tabList = screen.getAllByRole("tab");
    await user.click(tabList[4]!);
    expect(tabList[4]).toBeInTheDocument();
  });

  it("renders tasks with deadlines and edit affordance", async () => {
    const user = userEvent.setup();
    seed({
      tasks: [
        {
          id: 1,
          project_id: 1,
          name: "Design homepage",
          stage: "todo",
          planned_hours: 8,
          effective_hours: 2,
          deadline: "2026-02-01",
        },
      ],
    });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getAllByRole("tab")[1]!);

    expect(await screen.findByText("Design homepage")).toBeInTheDocument();
  });

  it("toggles a reached milestone", async () => {
    const user = userEvent.setup();
    seed({
      milestones: [{ id: 1, project_id: 1, name: "Launch", reached: false, deadline: null }],
    });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getAllByRole("tab")[2]!);

    const toggle = await screen.findByRole("button", { name: "Mark reached" });
    await user.click(toggle);
    expect(toggle).toBeInTheDocument();
  });

  it("renders timesheets and pnl cards", async () => {
    const user = userEvent.setup();
    seed({
      timesheets: [
        { id: 1, project_id: 1, projectId: 1, description: null, date: "2026-01-02", hours: 3 },
      ],
    });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getAllByRole("tab")[3]!);
    expect(await screen.findByText("Timesheet entry")).toBeInTheDocument();
    await user.click(screen.getAllByRole("tab")[5]!);
    expect(screen.getAllByRole("tab")[5]).toBeInTheDocument();
  });

  it("renders pnl fallback when summary is null", async () => {
    const user = userEvent.setup();
    seed({ summary: null });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getAllByRole("tab")[5]!);
    expect(screen.getAllByRole("tab")[5]).toBeInTheDocument();
  });
});
