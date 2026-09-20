import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { BankStatementsSection } from "../bank-statements-section";

function useLocalStatements() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/bank-statements", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          bank_statements: [
            {
              id: 1,
              journal_id: 1,
              name: "BS-January",
              date: "2026-01-31",
              balance_start: 10000,
              balance_end: 12500,
              state: "open",
              created_at: "2026-01-31T00:00:00Z",
              updated_at: "2026-01-31T00:00:00Z",
            },
            {
              id: 2,
              journal_id: 1,
              name: "BS-February",
              date: "2026-02-28",
              balance_start: 12500,
              balance_end: 14000,
              state: "draft",
              created_at: "2026-02-28T00:00:00Z",
              updated_at: "2026-02-28T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          journals: [
            {
              id: 1,
              organization_id: 1,
              name: "Bank Journal",
              code: "BNK",
              type: "bank",
              default_account_id: null,
              bank_account_id: null,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalStatements();
});

describe("BankStatementsSection", () => {
  it("renders seeded statements", async () => {
    renderWithProviders(<BankStatementsSection orgId="1" />);

    expect(await screen.findByText("BS-January")).toBeInTheDocument();
    expect(screen.getByText("BS-February")).toBeInTheDocument();
  });

  it("filters statements by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<BankStatementsSection orgId="1" />);

    await screen.findByText("BS-January");
    await user.type(screen.getByPlaceholderText("Search statements..."), "February");

    expect(await screen.findByText("BS-February")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<BankStatementsSection orgId="1" />);

    await screen.findByText("BS-January");
    await user.click(screen.getByRole("button", { name: "New statement" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Create bank statement")).toBeInTheDocument();
  });
});
