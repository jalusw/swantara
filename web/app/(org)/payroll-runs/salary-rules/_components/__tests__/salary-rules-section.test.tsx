import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SalaryRulesSection } from "../salary-rules-section";

function useLocalRules() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/salary-rules", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          salary_rules: [
            {
              id: 1,
              organization_id: 1,
              code: "BASIC",
              name: "Basic salary",
              category: "earning",
              compute_type: "fixed",
              amount: 5000,
              formula: null,
              account_debit_id: null,
              account_credit_id: null,
            },
            {
              id: 2,
              organization_id: 1,
              code: "TAX",
              name: "Income tax",
              category: "deduction",
              compute_type: "percent",
              amount: 10,
              formula: null,
              account_debit_id: null,
              account_credit_id: null,
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalRules();
});

describe("SalaryRulesSection", () => {
  it("renders seeded salary rules", async () => {
    renderWithProviders(<SalaryRulesSection orgId="1" />);

    expect(await screen.findByText("BASIC")).toBeInTheDocument();
    expect(screen.getByText("TAX")).toBeInTheDocument();
  });

  it("filters rules by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SalaryRulesSection orgId="1" />);

    await screen.findByText("BASIC");
    await user.type(screen.getByPlaceholderText("Search rules…"), "TAX");

    expect((await screen.findAllByText("TAX")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("BASIC")).not.toBeInTheDocument());
  });
});
