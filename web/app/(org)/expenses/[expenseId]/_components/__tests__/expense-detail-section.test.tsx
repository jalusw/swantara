import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ExpenseDetail } from "../expense-detail-section";

function useLocalReport() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/expense-reports/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          report: {
            id: 7,
            name: "Trip to Jakarta",
            employee_id: 1,
            state: "draft",
            payment_mode: "own_account",
            total_amount: 1500,
            entry_id: null,
            submitted_at: null,
            approved_by: null,
            lines: [
              {
                id: 31,
                category_id: 1,
                item_id: null,
                description: "Hotel stay",
                expense_date: "2026-03-01",
                quantity: 2,
                unit_price: 750,
                amount: 1500,
                tax_ids: [],
                dimension_id: null,
                project_id: null,
                reimbursable: true,
                receipt_attachment_id: null,
              },
            ],
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalReport();
});

describe("ExpenseDetail", () => {
  it("renders the report header with total", async () => {
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);

    expect((await screen.findAllByText("Trip to Jakarta")).length).toBeGreaterThan(0);
    expect(screen.getByText("1.500")).toBeInTheDocument();
  });

  it("switches to the lines tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);

    await screen.findAllByText("Trip to Jakarta");
    await user.click(screen.getByRole("tab", { name: "Baris" }));

    expect(await screen.findByText("Hotel stay")).toBeInTheDocument();
  });
});
