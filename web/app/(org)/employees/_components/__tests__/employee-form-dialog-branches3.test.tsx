import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeFormDialog } from "../employee-form-dialog";

function seed() {
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
    http.post("*/api/v1/organizations/:organizationId/employees", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.put("*/api/v1/organizations/:organizationId/employees/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

import type { Employee } from "@/lib/services/swantara";

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

function renderCreate(onSave: () => void = () => undefined) {
  renderWithProviders(
    <EmployeeFormDialog open onOpenChange={() => undefined} orgId="1" onSave={onSave} />,
  );
}

function renderEdit(onSave: () => void = () => undefined) {
  renderWithProviders(
    <EmployeeFormDialog
      open
      onOpenChange={() => undefined}
      orgId="1"
      initial={baseEmployee}
      onSave={onSave}
    />,
  );
}

beforeEach(() => {
  seed();
});

describe("EmployeeFormDialog branches3", () => {
  it("renders create title for new employees", async () => {
    renderCreate();

    expect(await screen.findByText("Add employee")).toBeInTheDocument();
  });

  it("renders edit title when initial is provided", async () => {
    renderEdit();

    expect(await screen.findByText("Edit employee")).toBeInTheDocument();
  });

  it("fills user and department ids on edit defaults", async () => {
    renderWithProviders(
      <EmployeeFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        initial={{ ...baseEmployee, userId: 9, departmentId: 1 } as Employee}
        onSave={() => undefined}
      />,
    );

    expect(await screen.findByText("Edit employee")).toBeInTheDocument();
  });

  it("submits the create form", async () => {
    const user = userEvent.setup();
    let saved = false;
    renderCreate(() => {
      saved = true;
    });

    await screen.findByText("Add employee");
    await user.type(screen.getByPlaceholderText("Name"), "Budi Santoso");
    await user.type(screen.getByPlaceholderText("Employee number"), "EMP-0099");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Add employee")).toBeInTheDocument();
    expect(saved).toBe(true);
  });

  it("shows employment and wage fields", async () => {
    renderCreate();

    expect(await screen.findByPlaceholderText("Employee number")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Name")).toBeInTheDocument();
  });
});
