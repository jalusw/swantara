import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Employee } from "@/lib/services/swantara/types";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeFormDialog } from "../employee-form-dialog";

function seedRefs() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { job_positions: [] } }),
    ),
  );
}

const baseEmployee = {
  id: 3,
  organizationId: 1,
  contactId: 1,
  userId: null,
  employeeNumber: "EMP-0003",
  departmentId: null,
  jobPositionId: null,
  managerId: null,
  hireDate: null,
  terminationDate: null,
  employmentType: "full_time",
  workLocation: null,
  active: true,
} as unknown as Employee;

beforeEach(() => {
  seedRefs();
});

describe("EmployeeFormDialog branches2", () => {
  it("renders the edit title branch with initial employee", () => {
    renderWithProviders(
      <EmployeeFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={baseEmployee}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByText("Ubah karyawan")).toBeInTheDocument();
  });

  it("creates an employee with null optionals branch", async () => {
    let createBody: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/employees", async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} }, { status: 201 });
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.type(screen.getByLabelText("Nama"), "Ayu Lestari");
    await user.type(screen.getByLabelText("Nomor karyawan"), "EMP-0099");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(createBody).toMatchObject({ employee_number: "EMP-0099", user_id: null });
  });

  it("updates an employee through the edit branch with empty optionals", async () => {
    let updateBody: unknown = null;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/employees/:id", async ({ request }) => {
        updateBody = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <EmployeeFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={baseEmployee}
        onSave={onSave}
      />,
    );

    await user.type(screen.getByLabelText("Nama"), "Budi Santoso");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(updateBody).toMatchObject({ department_id: null, job_position_id: null });
  });

  it("shows the create title branch when no initial is given", () => {
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(screen.getByText("Tambah Karyawan")).toBeInTheDocument();
  });
});
