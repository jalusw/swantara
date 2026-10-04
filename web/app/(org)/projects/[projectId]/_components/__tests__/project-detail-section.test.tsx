import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProjectDetail } from "../project-detail-section";

function useLocalProject() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/projects/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          project: {
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
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:id/summary", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          summary: {
            planned_hours: 100,
            effective_hours: 40,
            billable_amount: 4000,
            billed_amount: 1000,
            cost: 2000,
            revenue: 4000,
          },
        },
      }),
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
              deadline: null,
              priority: null,
            },
          ],
        },
      }),
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

beforeEach(() => {
  useLocalProject();
});

describe("ProjectDetail", () => {
  it("renders project name", async () => {
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    expect(await screen.findByRole("heading", { name: "Website Redesign" })).toBeInTheDocument();
  });

  it("switches to tasks tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProjectDetail orgId="1" projectId="1" />);

    await screen.findByRole("heading", { name: "Website Redesign" });
    await user.click(screen.getByRole("tab", { name: "Tugas" }));

    expect(await screen.findByText("Design homepage")).toBeInTheDocument();
  });
});
