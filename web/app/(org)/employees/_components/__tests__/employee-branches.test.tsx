import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeFormDialog } from "../employee-form-dialog";

const departments = [{ id: 2, name: "Engineering" }];
const positions = [{ id: 3, name: "Developer" }];

function seedLists() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { job_positions: positions } }),
    ),
  );
}

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/employees", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { employee: { id: 9 } },
      });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/employees/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { employee: { id: 4 } },
      });
    }),
  );
}

const editInitial = {
  id: 4,
  employeeNumber: "EMP-0004",
  userId: null,
  departmentId: null,
  jobPositionId: null,
  managerId: null,
  hireDate: null,
  terminationDate: null,
  employmentType: "full_time",
  workLocation: null,
  active: true,
} as never;

const editInitialFilled = {
  id: 5,
  employeeNumber: "EMP-0005",
  userId: 7,
  departmentId: 2,
  jobPositionId: 3,
  managerId: 1,
  hireDate: "2024-02-01",
  terminationDate: null,
  employmentType: "part_time",
  workLocation: "Jakarta",
  active: true,
} as never;

beforeEach(() => {
  seedLists();
});

describe("EmployeeFormDialog branches", () => {
  it("shows edit title with prefilled employee number", async () => {
    renderWithProviders(
      <EmployeeFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Ubah karyawan")).toBeInTheDocument();
    expect(screen.getByDisplayValue("EMP-0004")).toBeInTheDocument();
  });

  it("blocks submit when required fields are empty", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).not.toHaveBeenCalled());
    expect(screen.getByDisplayValue("0")).toBeInTheDocument();
  });

  it("creates an employee with contract and hourly terms", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Alex");
    await user.type(screen.getByLabelText("Nomor karyawan"), "EMP-1");
    await user.click(screen.getByRole("combobox", { name: "Jenis kepegawaian" }));
    await user.click(await screen.findByRole("option", { name: "Kontrak" }));
    await user.click(screen.getByRole("combobox", { name: "Jenis upah" }));
    await user.click(await screen.findByRole("option", { name: "Per jam" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates an employee with only required fields", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Alex");
    await user.type(screen.getByLabelText("Nomor karyawan"), "EMP-1");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates an employee with optional fields filled", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "June");
    await user.type(screen.getByLabelText("Nomor karyawan"), "EMP-2");
    await user.type(screen.getByPlaceholderText("Email"), "june@acme.com");
    await user.type(screen.getByPlaceholderText("Phone"), "+62 811");
    await user.click(screen.getByRole("combobox", { name: "Departemen" }));
    await user.click(await screen.findByRole("option", { name: "Engineering" }));
    await user.click(screen.getByRole("combobox", { name: "Jabatan" }));
    await user.click(await screen.findByRole("option", { name: "Developer" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Alex");
    await user.type(screen.getByLabelText("Nomor karyawan"), "EMP-1");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Tambah Karyawan")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates an employee with empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <EmployeeFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Alex Updated");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates an employee with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <EmployeeFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah karyawan");
    expect(screen.getByDisplayValue("Jakarta")).toBeInTheDocument();
    await user.type(screen.getByPlaceholderText("Nama"), "June Updated");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <EmployeeFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah karyawan");
    await user.type(screen.getByPlaceholderText("Nama"), "Alex");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Ubah karyawan")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <EmployeeFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Tambah Karyawan");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
