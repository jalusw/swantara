import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { StatementDetailSection } from "../statement-detail-section";

function useLocalStatement() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/bank-statements/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          bank_statement: {
            id: 5,
            journal_id: 1,
            name: "BS-March",
            date: "2026-03-31",
            balance_start: 14000,
            balance_end: 14000,
            state: "open",
            created_at: "2026-03-31T00:00:00Z",
            updated_at: "2026-03-31T00:00:00Z",
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalStatement();
});

describe("StatementDetailSection", () => {
  it("renders statement balances", async () => {
    renderWithProviders(<StatementDetailSection orgId="1" statementId="5" />);

    expect(await screen.findByText("BS-March")).toBeInTheDocument();
    expect(screen.getAllByText("14,000.00").length).toBeGreaterThan(0);
  });

  it("renders empty lines state", async () => {
    renderWithProviders(<StatementDetailSection orgId="1" statementId="5" />);

    await screen.findByText("BS-March");
    expect(screen.getByText("Statement lines")).toBeInTheDocument();
  });
});
