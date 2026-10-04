import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import type { Employee } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeFormDialog } from "../employee-form-dialog";

const created: unknown[] = [];
const updated: unknown[] = [];

function seed() {
  created.length = 0;
  updated.length = 0;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { departments: [{ id: 1, name: "Engineering" }] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { job_positions: [{ id: 2, name: "Developer" }] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/employees", async ({ request }) => {
      created.push(await request.json());
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
    http.put("*/api/v1/organizations/:organizationId/employees/:id", async ({ request }) => {
      updated.push(await request.json());
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
  );
}

const baseEmployee = {
  id: 1,
  employeeNumber: "EMP-0001",
  employee_number: "EMP-0001",
  userId: null,
  user_id: null,
  departmentId: null,
  department_id: null,
  jobPositionId: null,
  job_position_id: null,
  managerId: null,
  manager_id: null,
  hireDate: null,
  hire_date: null,
  terminationDate: null,
  termination_date: null,
  employmentType: "full_time",
  employment_type: "full_time",
  workLocation: null,
  work_location: null,
  active: true,
} as unknown as Employee;

beforeEach(() => {
  seed();
});

describe("EmployeeFormDialog branches4", () => {
  it("submits an update with zeroed ids mapped to null", async () => {
    const user = userEvent.setup();
    let saved = false;
    renderWithProviders(
      <EmployeeFormDialog
        open
        onOpenChange={() => undefined}
        orgId=""
        initial={{ ...baseEmployee, userId: 0, departmentId: 0, jobPositionId: 0 } as Employee}
        onSave={() => {
          saved = true;
        }}
      />,
    );

    await screen.findByText("Ubah karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Budi Santoso");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe(true));
    expect(updated.length).toBe(1);
    const body = updated[0] as Record<string, unknown>;
    expect(body.user_id).toBeNull();
    expect(body.department_id).toBeNull();
    expect(body.job_position_id).toBeNull();
  });

  it("submits a create with edited and cleared wage", async () => {
    const user = userEvent.setup();
    let saved = false;
    renderWithProviders(
      <EmployeeFormDialog
        open
        onOpenChange={() => undefined}
        orgId=""
        onSave={() => {
          saved = true;
        }}
      />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Siti Rahayu");
    await user.type(screen.getByPlaceholderText("Nomor karyawan"), "EMP-0101");
    const wage = screen.getByPlaceholderText("0");
    await user.type(wage, "7500");
    await user.clear(wage);
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe(true));
    expect(created.length).toBe(1);
    const body = created[0] as Record<string, unknown>;
    expect(body.organization_id).toBe(0);
    expect(body.wage).toBe(0);
  });
});
