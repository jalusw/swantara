import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TimesheetFormDialog } from "../timesheet-form-dialog";

const employees = [
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
];
const contacts = [{ id: 10, organization_id: 1, name: "Alex Rivera", display_name: null }];
const accounts = [{ id: 21, name: "Operating costs" }];
const projects = [{ id: 31, name: "Website Revamp" }];
const tasks = [{ id: 41, name: "Design homepage" }];

function seedLists() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/employees", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { employees } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/dimensions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { projects } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects/:projectId/tasks", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { tasks } }),
    ),
  );
}

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/timesheets", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { timesheet: { id: 9 } },
      });
    }),
  );
}

beforeEach(() => {
  seedLists();
});

describe("TimesheetFormDialog branches", () => {
  it("blocks submit when no employee is selected", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Add timesheet entry");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Select an employee.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates an entry with only an employee", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Add timesheet entry");
    await user.click(screen.getByRole("combobox", { name: "Employee" }));
    await user.click(await screen.findByRole("option", { name: "Alex Rivera" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("enables task selection after shipment a project", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Add timesheet entry");
    expect(screen.getByRole("combobox", { name: "Task" })).toBeDisabled();
    await user.click(screen.getByRole("combobox", { name: "Employee" }));
    await user.click(await screen.findByRole("option", { name: "Alex Rivera" }));
    await user.click(screen.getByRole("combobox", { name: "Project" }));
    await user.click(await screen.findByRole("option", { name: "Website Revamp" }));

    const task = screen.getByRole("combobox", { name: "Task" });
    await waitFor(() => expect(task).toBeEnabled());
    await user.click(task);
    await user.click(await screen.findByRole("option", { name: "Design homepage" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates an entry with description and dimension account", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Add timesheet entry");
    await user.click(screen.getByRole("combobox", { name: "Employee" }));
    await user.click(await screen.findByRole("option", { name: "Alex Rivera" }));
    await user.type(
      screen.getByLabelText("Log hours worked on projects and tasks."),
      "Worked on homepage",
    );
    await user.click(screen.getByRole("combobox", { name: "Dimension account" }));
    await user.click(await screen.findByRole("option", { name: "Operating costs" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("blocks submit when hours fall below the minimum", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Add timesheet entry");
    await user.click(screen.getByRole("combobox", { name: "Employee" }));
    await user.click(await screen.findByRole("option", { name: "Alex Rivera" }));
    const hours = screen.getByLabelText("Hours");
    await user.clear(hours);
    await user.type(hours, "0");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Minimum 0.5 hours.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Add timesheet entry");
    await user.click(screen.getByRole("combobox", { name: "Employee" }));
    await user.click(await screen.findByRole("option", { name: "Alex Rivera" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(screen.getByText("Add timesheet entry")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Add timesheet entry");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
