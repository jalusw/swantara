import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ExpenseFormDialog } from "../expense-form-dialog";

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
    http.get("*/api/v1/organizations/:organizationId/expense-categories", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          categories: [
            {
              id: 1,
              organization_id: 1,
              name: "Travel",
              expense_account_id: null,
              default_tax_ids: [],
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalForm();
});

describe("ExpenseFormDialog", () => {
  it("renders the create title with name field", async () => {
    renderWithProviders(
      <ExpenseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Klaim biaya baru")).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("accepts a report name", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <ExpenseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const nameInput = await screen.findByLabelText("Nama");
    await user.type(nameInput, "Trip to Jakarta");

    expect(nameInput).toHaveValue("Trip to Jakarta");
  });
});
