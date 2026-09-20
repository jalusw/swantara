import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeaveRequestFormDialog } from "../leave-request-form-dialog";

function useLocalForm() {
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
  useLocalForm();
});

describe("LeaveRequestFormDialog", () => {
  it("renders create title with date fields", async () => {
    renderWithProviders(
      <LeaveRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Add leave request")).toBeInTheDocument();
  });

  it("renders days input", async () => {
    renderWithProviders(
      <LeaveRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Add leave request");
    expect(screen.getByLabelText("Days")).toBeInTheDocument();
  });
});
