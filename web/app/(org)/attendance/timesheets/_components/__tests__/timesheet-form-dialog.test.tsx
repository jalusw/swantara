import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TimesheetFormDialog } from "../timesheet-form-dialog";

function useLocalOptions() {
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
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/dimensions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/projects", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { projects: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalOptions();
});

describe("TimesheetFormDialog", () => {
  it("renders create title with hours field", async () => {
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Tambah entri lembar waktu")).toBeInTheDocument();
    expect(screen.getByLabelText("Jam")).toBeInTheDocument();
  });

  it("renders employee options from the API", async () => {
    renderWithProviders(
      <TimesheetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Tambah entri lembar waktu");
    expect(screen.getAllByText("Karyawan").length).toBeGreaterThan(0);
  });
});
