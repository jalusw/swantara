import { screen, waitFor } from "@testing-library/react";
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
    manager_id: 10,
    dimension_id: null,
    sale_order_id: null,
    billing_type: "fixed",
    billable_rate: 100,
    date_start: "2026-01-05",
    date_end: "2026-06-30",
    state: "open",
    ...overrides,
  };
}

function summaryFixture(overrides = {}) {
  return {
    project_id: 1,
    planned_hours: 100,
    effective_hours: 40,
    utilization: 0.4,
    billed_amount: 1000,
    unbilled_amount: 3000,
    billable_amount: 4000,
    cost_amount: 2000,
    margin_amount: 2000,
    ...overrides,
  };
}

function useProjectHandlers(
  project: Record<string, unknown>,
  opts?: { summary?: Record<string, unknown> | null },
) {
  const summary = opts?.summary === undefined ? summaryFixture() : opts.summary;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { project } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:id/summary", () =>
      summary
        ? HttpResponse.json({ success: true, message: "OK.", data: { summary } })
        : HttpResponse.json({ success: false, message: "No summary." }, { status: 404 }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/tasks", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
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
              deadline: "2026-04-01",
              priority: null,
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/milestones", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          milestones: [
            {
              id: 5,
              project_id: 1,
              name: "Launch",
              deadline: null,
              reached: false,
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/timesheets", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          timesheets: [
            {
              id: 3,
              project_id: 1,
              description: "Homepage mockups",
              date: "2026-03-02",
              hours: 4,
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

describe("ProjectDetail remainder", () => {
  it("shows the not-found state when the project is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<ProjectDetail orgId="1" projectId="999" />);

    expect(await screen.findByText("Proyek tidak ditemukan.")).toBeInTheDocument();
  });

  it("renders manager, dates and summary cards on the overview tab", async () => {
    useProjectHandlers(projectFixture());
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    expect(await screen.findByRole("heading", { name: "Website Redesign" })).toBeInTheDocument();
    expect(screen.getAllByText("Acme Corp").length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText("Jam rencana")).toBeInTheDocument();
  });

  it("closes an open project from the header action", async () => {
    let closedState: unknown = null;
    useProjectHandlers(projectFixture());
    server.use(
      http.put("*/api/v1/organizations/:organizationId/projects/:id/state", async ({ request }) => {
        closedState = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("button", { name: "Tutup" }));

    await waitFor(() => expect(closedState).toMatchObject({ state: "closed" }));
  });

  it("opens a draft project from the header action", async () => {
    let openedState: unknown = null;
    useProjectHandlers(projectFixture({ state: "draft" }));
    server.use(
      http.put("*/api/v1/organizations/:organizationId/projects/:id/state", async ({ request }) => {
        openedState = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("button", { name: "Buka" }));

    await waitFor(() => expect(openedState).toMatchObject({ state: "open" }));
  });

  it("shows task deadlines and opens the edit task dialog", async () => {
    const user = userEvent.setup();
    useProjectHandlers(projectFixture());
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tugas" }));

    expect(await screen.findByText("Design homepage")).toBeInTheDocument();
    expect(screen.getByText(/due/i)).toBeInTheDocument();
  });

  it("toggles a milestone to reached", async () => {
    let reached: unknown = null;
    useProjectHandlers(projectFixture());
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/projects/:projectId/milestones/:id/reached",
        async ({ request }) => {
          reached = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tonggak" }));

    expect(await screen.findByText("Launch")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Tandai tercapai" }));

    await waitFor(() => expect(reached).toMatchObject({ reached: true }));
  });

  it("shows timesheets on the timesheets tab", async () => {
    const user = userEvent.setup();
    useProjectHandlers(projectFixture());
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Lembar waktu" }));

    expect(await screen.findByText("Homepage mockups")).toBeInTheDocument();
  });

  it("shows the non-time-material message on the billing tab", async () => {
    const user = userEvent.setup();
    useProjectHandlers(projectFixture({ billing_type: "fixed" }));
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(
      await screen.findByText("Penagihan hanya tersedia untuk proyek Waktu & Material."),
    ).toBeInTheDocument();
  });

  it("shows pnl cards on the pnl tab and no-data without a summary", async () => {
    const user = userEvent.setup();
    useProjectHandlers(projectFixture(), { summary: null });
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Laba rugi" }));

    expect(await screen.findByText("Tidak ada data keuangan yang tersedia.")).toBeInTheDocument();
  });

  it("opens the edit project dialog from the header", async () => {
    const user = userEvent.setup();
    useProjectHandlers(projectFixture());
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("button", { name: "Ubah proyek" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
