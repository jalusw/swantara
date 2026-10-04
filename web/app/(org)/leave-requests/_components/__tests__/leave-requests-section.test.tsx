import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeaveRequestsSection } from "../leave-requests-section";

function useLocalLeave() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/leave-requests", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          leave_requests: [
            {
              id: 1,
              employee_id: 1,
              leave_type_id: 1,
              date_from: "2026-03-01",
              date_to: "2026-03-03",
              days: 3,
              state: "submitted",
            },
            {
              id: 2,
              employee_id: 2,
              leave_type_id: 1,
              date_from: "2026-04-01",
              date_to: "2026-04-02",
              days: 2,
              state: "draft",
            },
          ],
        },
      }),
    ),
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
              hire_date: "2024-01-01",
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
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
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
            { id: 10, organization_id: 1, name: "Alex Rivera", display_name: "Alex Rivera" },
            { id: 11, organization_id: 1, name: "June Park", display_name: "June Park" },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/leave-types", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          leave_types: [
            {
              id: 1,
              organization_id: 1,
              name: "Annual Leave",
              paid: true,
              allocation_days: 12,
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalLeave();
});

describe("LeaveRequestsSection", () => {
  it("renders seeded leave requests", async () => {
    renderWithProviders(<LeaveRequestsSection orgId="1" />);

    expect(await screen.findByText("Alex Rivera")).toBeInTheDocument();
    expect(screen.getAllByText("Annual Leave").length).toBeGreaterThan(0);
  });

  it("offers approve action for submitted requests", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeaveRequestsSection orgId="1" />);

    expect(await screen.findByRole("button", { name: "Setujui" })).toBeInTheDocument();
    await user.hover(screen.getByText("Alex Rivera"));
    expect(screen.getByText("June Park")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeaveRequestsSection orgId="1" />);

    await screen.findByText("Alex Rivera");
    await user.click(screen.getByRole("button", { name: "Tambah pengajuan" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
