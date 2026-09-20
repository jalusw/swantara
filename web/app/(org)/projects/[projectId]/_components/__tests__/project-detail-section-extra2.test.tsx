import { screen, waitFor, within } from "@testing-library/react";
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
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { summary: summaryOverride },
      }),
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

describe("ProjectDetail extra2", () => {
  it("opens a draft project from the header actions", async () => {
    seedProject({ project: { state: "draft" } });
    const stateCalls: unknown[] = [];
    server.use(
      http.put("*/api/v1/organizations/:organizationId/projects/:id/state", async ({ request }) => {
        stateCalls.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: { project } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("button", { name: "Open" }));

    await waitFor(() => expect(stateCalls).toHaveLength(1));
    expect(stateCalls[0]).toMatchObject({ state: "open" });
  });

  it("cancels an open project from the header actions", async () => {
    seedProject();
    const stateCalls: unknown[] = [];
    server.use(
      http.put("*/api/v1/organizations/:organizationId/projects/:id/state", async ({ request }) => {
        stateCalls.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: { project } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(stateCalls).toHaveLength(1));
    expect(stateCalls[0]).toMatchObject({ state: "cancelled" });
  });

  it("opens the add-task dialog from the tasks tab", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tasks" }));
    await user.click(await screen.findByRole("button", { name: "New task" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("opens the edit-task dialog for an existing task", async () => {
    seedProject({
      tasks: [
        {
          id: 1,
          project_id: 1,
          name: "Design homepage",
          assignee_id: null,
          stage: "todo",
          planned_hours: 20,
          effective_hours: 5,
          parent_task_id: null,
          deadline: null,
          priority: null,
        },
      ],
    });
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tasks" }));
    await screen.findByText("Design homepage");
    const panel = screen
      .getAllByRole("tabpanel")
      .find((candidate) => !candidate.hasAttribute("inert"))!;
    const iconButtons = within(panel)
      .getAllByRole("button")
      .filter((button) => button.textContent === "");
    expect(iconButtons).toHaveLength(1);
    await user.click(iconButtons[0]!);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("shows the task deadline when present", async () => {
    seedProject({
      tasks: [
        {
          id: 1,
          project_id: 1,
          name: "Design homepage",
          assignee_id: null,
          stage: "todo",
          planned_hours: 20,
          effective_hours: 5,
          parent_task_id: null,
          deadline: "2026-03-01",
          priority: null,
        },
      ],
    });
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tasks" }));

    expect(await screen.findByText(/Due/)).toBeInTheDocument();
  });

  it("opens the add-milestone dialog from the milestones tab", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Milestones" }));
    await user.click(await screen.findByRole("button", { name: "New milestone" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("marks a reached milestone as unreached", async () => {
    seedProject({
      milestones: [{ id: 4, project_id: 1, name: "Go live", reached: true, deadline: null }],
    });
    const reachedBodies: unknown[] = [];
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/projects/:projectId/milestones/:milestoneId/reached",
        async ({ request }) => {
          reachedBodies.push(await request.json());
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Milestones" }));
    expect(await screen.findByText("Reached")).toBeInTheDocument();
    await user.click(await screen.findByRole("button", { name: "Mark unreached" }));

    await waitFor(() => expect(reachedBodies).toHaveLength(1));
    expect(reachedBodies[0]).toMatchObject({ reached: false });
  });

  it("shows the empty tasks state when no tasks exist", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tasks" }));

    expect(await screen.findByText("No tasks yet.")).toBeInTheDocument();
  });

  it("shows the empty timesheets state when no timesheets exist", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Timesheets" }));

    expect(await screen.findByText("No tasks yet.")).toBeInTheDocument();
  });

  it("shows the P&L no-data state when the summary is missing", async () => {
    seedProject({ summary: null });
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "P&L" }));

    expect(await screen.findByText("No financial data available.")).toBeInTheDocument();
  });
});
