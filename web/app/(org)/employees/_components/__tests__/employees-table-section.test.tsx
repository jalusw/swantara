import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeesTable } from "../employees-table-section";

function useLocalEmployees() {
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
            {
              id: 2,
              organization_id: 1,
              contact_id: 11,
              employee_number: "EMP-0002",
              department_id: null,
              job_position_id: null,
              manager_id: null,
              hire_date: "2024-02-01",
              termination_date: null,
              employment_type: "contract",
              work_location: "Bandung",
              active: false,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
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
  useLocalEmployees();
});

describe("EmployeesTable", () => {
  it("renders seeded employees", async () => {
    renderWithProviders(<EmployeesTable orgId="1" />);

    expect(await screen.findByText("EMP-0001")).toBeInTheDocument();
    expect(screen.getByText("EMP-0002")).toBeInTheDocument();
  });

  it("filters employees by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EmployeesTable orgId="1" />);

    await screen.findByText("EMP-0001");
    await user.type(screen.getByPlaceholderText("Cari karyawan…"), "0002");

    expect(await screen.findByText("EMP-0002")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EmployeesTable orgId="1" />);

    await screen.findByText("EMP-0001");
    await user.click(screen.getByRole("button", { name: "Tambah Karyawan" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
