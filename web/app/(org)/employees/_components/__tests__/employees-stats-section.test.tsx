import { screen, within } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeesStats } from "../employees-stats-section";

function useLocalStats() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/employees", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          employees: [
            {
              id: 1,
              organization_id: 1,
              contact_id: 10,
              employee_number: "EMP-0001",
              hire_date: "2026-01-05",
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              contact_id: 11,
              employee_number: "EMP-0002",
              hire_date: "2024-01-01",
              active: false,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/leave-requests", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { leave_requests: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalStats();
});

describe("EmployeesStats", () => {
  it("renders stat labels", async () => {
    renderWithProviders(<EmployeesStats />);

    expect(await screen.findByText("Total employees")).toBeInTheDocument();
    expect(screen.getByText("Active employees")).toBeInTheDocument();
    expect(screen.getByText("On leave")).toBeInTheDocument();
    expect(screen.getByText("New this month")).toBeInTheDocument();
  });

  it("renders total headcount", async () => {
    renderWithProviders(<EmployeesStats />);

    await screen.findByText("Total employees");
    const grid = screen
      .getByText("Total employees")
      .closest('[data-slot="metric-grid"]') as HTMLElement;
    const total = within(grid).getByText("Total employees").closest('[data-slot="stat-card"]');
    expect(await within(total as HTMLElement).findByText("2")).toBeInTheDocument();
  });
});
