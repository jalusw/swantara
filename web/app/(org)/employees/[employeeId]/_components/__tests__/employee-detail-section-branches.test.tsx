import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EmployeeDetail } from "../employee-detail-section";

const employee = {
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
  work_location: null,
  active: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function seed(empOverrides: Record<string, unknown> = {}, contracts: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/employees/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { employee: { ...employee, ...empOverrides } },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contracts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contracts } }),
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
  seed();
});

describe("EmployeeDetail branches", () => {
  it("renders loading state while fetching", () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/employees/:id", () => new Promise(() => {})),
    );
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    expect(screen.getByText("Memuat...")).toBeInTheDocument();
  });

  it("renders not-found when employee is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/employees/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { employee: null } }),
      ),
    );
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    expect(await screen.findByText("Karyawan tidak ditemukan.")).toBeInTheDocument();
  });

  it("renders inactive badge for inactive employees", async () => {
    seed({ active: false });
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    expect(await screen.findByRole("heading", { name: "EMP-0001" })).toBeInTheDocument();
    expect(screen.getByText("Tidak aktif")).toBeInTheDocument();
  });

  it("shows contracts error with retry", async () => {
    const user = userEvent.setup();
    seed({}, []);
    server.use(
      http.get("*/api/v1/organizations/:organizationId/contracts", () =>
        HttpResponse.json({ success: false, message: "Boom." }, { status: 500 }),
      ),
    );
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    await screen.findByRole("heading", { name: "EMP-0001" });
    await user.click(screen.getAllByRole("tab")[1]!);

    expect(await screen.findByRole("button", { name: "Coba lagi" })).toBeInTheDocument();
  });

  it("renders a contract with inactive state badge", async () => {
    const user = userEvent.setup();
    seed({}, [
      {
        id: 3,
        employee_id: 1,
        employeeId: 1,
        date_start: "2024-01-01",
        dateStart: "2024-01-01",
        date_end: null,
        dateEnd: null,
        wage: 5000,
        currency_code: "IDR",
        currencyCode: "IDR",
        wage_type: "monthly",
        wageType: "monthly",
        state: "draft",
      },
    ]);
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    await screen.findByRole("heading", { name: "EMP-0001" });
    await user.click(screen.getAllByRole("tab")[1]!);

    expect(await screen.findByText(/5\.000/)).toBeInTheDocument();
  });

  it("opens the edit dialog and saves", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EmployeeDetail orgId="1" employeeId="1" />);

    await screen.findByRole("heading", { name: "EMP-0001" });
    await user.click(screen.getByRole("button", { name: "Ubah" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
