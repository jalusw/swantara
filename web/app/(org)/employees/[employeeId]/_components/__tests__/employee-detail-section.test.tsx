import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeDetail } from "../employee-detail-section";

function useLocalEmployee() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/employees/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          employee: {
            id: 1,
            organization_id: 1,
            contact_id: 10,
            user_id: null,
            employee_number: "EMP-0001",
            department_id: null,
            job_position_id: null,
            manager_id: null,
            hire_date: "2024-01-01",
            termination_date: null,
            employment_type: "full_time",
            work_location: "Jakarta",
            active: true,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contracts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contracts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { job_positions: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalEmployee();
});

describe("EmployeeDetail", () => {
  it("renders employee number", async () => {
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    expect(await screen.findByRole("heading", { name: "EMP-0001" })).toBeInTheDocument();
  });

  it("switches to contracts tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    await screen.findByRole("heading", { name: "EMP-0001" });
    await user.click(screen.getByRole("tab", { name: "Kontrak" }));

    expect(await screen.findByText("Tidak ada kontrak")).toBeInTheDocument();
  });
});
