import { screen, waitFor } from "@testing-library/react";
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

describe("ProjectDetail extra", () => {
  it("closes an open project from the header actions", async () => {
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
    await user.click(screen.getByRole("button", { name: "Tutup" }));

    await waitFor(() => expect(stateCalls).toHaveLength(1));
    expect(stateCalls[0]).toMatchObject({ state: "closed" });
  });

  it("opens the edit dialog from the header", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("button", { name: "Ubah proyek" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("shows the empty milestones state", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tonggak" }));

    expect(await screen.findByText("Belum ada tonggak.")).toBeInTheDocument();
  });

  it("toggles a milestone to reached", async () => {
    seedProject({
      milestones: [{ id: 4, project_id: 1, name: "Go live", reached: false, deadline: null }],
    });
    let reachedCalls = 0;
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/projects/:projectId/milestones/:milestoneId/reached",
        () => {
          reachedCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tonggak" }));
    await user.click(await screen.findByRole("button", { name: "Tandai tercapai" }));

    await waitFor(() => expect(reachedCalls).toBe(1));
  });

  it("lists project timesheets on the timesheets tab", async () => {
    seedProject({
      timesheets: [
        {
          id: 9,
          project_id: 1,
          projectId: 1,
          description: "Homepage layout",
          date: "2026-02-01",
          hours: 5,
        },
      ],
    });
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Lembar waktu" }));

    expect(await screen.findByText("Homepage layout")).toBeInTheDocument();
  });

  it("explains billing limits for fixed-price projects", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(
      await screen.findByText("Penagihan hanya tersedia untuk proyek Waktu & Material."),
    ).toBeInTheDocument();
  });

  it("shows unbilled amounts for time-and-material projects", async () => {
    seedProject({ project: { billing_type: "time_material" } });
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(await screen.findByText("Jam belum ditagih")).toBeInTheDocument();
    expect(screen.getByText("Jumlah belum ditagih")).toBeInTheDocument();
  });

  it("shows financial cards on the P&L tab", async () => {
    seedProject();
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Laba rugi" }));

    expect(await screen.findByText("Pendapatan")).toBeInTheDocument();
    expect(screen.getByText("Jam rencana vs jam efektif")).toBeInTheDocument();
  });

  it("shows the not-found state when the project is missing", async () => {
    seedProject();
    server.use(
      http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    expect(await screen.findByText("Proyek tidak ditemukan.")).toBeInTheDocument();
  });
});
