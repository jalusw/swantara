import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ExpensesSection } from "../expenses-section";

function useLocalExpenses() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/expense-reports", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          reports: [
            {
              id: 1,
              name: "Trip to Jakarta",
              employee_id: 1,
              state: "draft",
              payment_mode: "own_account",
              total_amount: 1500,
              entry_id: null,
              submitted_at: null,
              approved_by: null,
            },
            {
              id: 2,
              name: "Team dinner",
              employee_id: 2,
              state: "submitted",
              payment_mode: "organization_account",
              total_amount: 800,
              entry_id: null,
              submitted_at: "2026-03-02T00:00:00Z",
              approved_by: null,
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalExpenses();
});

describe("ExpensesSection", () => {
  it("renders seeded expense reports", async () => {
    renderWithProviders(<ExpensesSection orgId="1" />);

    expect(await screen.findByText("Trip to Jakarta")).toBeInTheDocument();
    expect(screen.getByText("Team dinner")).toBeInTheDocument();
  });

  it("filters reports by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ExpensesSection orgId="1" />);

    await screen.findByText("Trip to Jakarta");
    await user.type(screen.getByPlaceholderText("Cari biaya"), "dinner");

    expect((await screen.findAllByText("Team dinner")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("Trip to Jakarta")).not.toBeInTheDocument());
  });
});
