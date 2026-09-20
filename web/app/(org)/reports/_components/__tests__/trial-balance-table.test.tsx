import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import { TrialBalanceTable } from "../trial-balance-table";

vi.mock("next/navigation", async (importOriginal) => ({
  ...((await importOriginal()) as Record<string, unknown>),
  useParams: () => ({ id: "1" }),
  useRouter: () => navigationMock,
  usePathname: () => "/",
}));

const rows = [
  {
    account_id: 10,
    code: "1000",
    name: "Cash",
    account_type: "asset",
    opening_debit: 500,
    opening_credit: 0,
    period_debit: 200,
    period_credit: 0,
    closing_debit: 700,
    closing_credit: 0,
  },
  {
    account_id: 40,
    code: "4000",
    name: "Revenue",
    account_type: "income",
    opening_debit: 0,
    opening_credit: 0,
    period_debit: 0,
    period_credit: 300,
    closing_debit: 0,
    closing_credit: 300,
  },
];

const trialBalance = {
  organization_id: 1,
  period_id: 2,
  start: "2026-01-01",
  end: "2026-01-31",
  rows,
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/reports/trial-balance", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { trialBalance: trialBalance } }),
    ),
  );
});

describe("TrialBalanceTable", () => {
  it("renders seeded rows with totals", async () => {
    renderWithProviders(<TrialBalanceTable />);

    expect(await screen.findByText("Cash")).toBeInTheDocument();
    expect(screen.getByText("Revenue")).toBeInTheDocument();
    expect(screen.getByText("asset")).toBeInTheDocument();
    expect(screen.getAllByText("IDR 500.00").length).toBeGreaterThan(0);
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);

    const link = screen.getByRole("link", { name: "1000" });
    expect(link).toHaveAttribute("href", "/accounts/10");
  });

  it("keeps row content visible on hover over the account link", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TrialBalanceTable />);

    const link = await screen.findByRole("link", { name: "1000" });
    await user.hover(link);

    expect(screen.getByText("Cash")).toBeInTheDocument();
    expect(link).toHaveAttribute("href", "/accounts/10");
  });

  it("shows the empty state when there are no rows", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/reports/trial-balance", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: { trialBalance: { ...trialBalance, rows: [] } },
        }),
      ),
    );
    renderWithProviders(<TrialBalanceTable />);

    expect(await screen.findByText("No trial balance data available.")).toBeInTheDocument();
  });
});
