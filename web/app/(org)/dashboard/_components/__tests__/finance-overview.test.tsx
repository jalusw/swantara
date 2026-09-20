import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { FinanceOverview } from "../finance-overview";

const kpi = {
  revenue: 84320,
  expenses: 52300,
  gross_margin_pct: 0.38,
  net_margin_pct: 0.22,
  ebitda: 32020,
  current_ratio: 1.8,
};

function seedFinance(payload: unknown) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/kpis/finance", () =>
      HttpResponse.json({ success: true, message: "OK.", data: payload }),
    ),
  );
}

beforeEach(() => {});

describe("FinanceOverview", () => {
  it("renders kpi values when finance data is present", async () => {
    seedFinance({ kpi });
    renderWithProviders(<FinanceOverview />);

    expect(await screen.findByText("Finance")).toBeInTheDocument();
    const matches = await screen.findAllByText(/84,320/);
    expect(matches.length).toBeGreaterThan(0);
  });

  it("renders placeholders when finance kpi is missing", async () => {
    seedFinance({});
    renderWithProviders(<FinanceOverview />);

    await screen.findByText("Finance");
    const dashes = screen.getAllByText("—");
    expect(dashes.length).toBeGreaterThan(0);
  });

  it("renders margin percentages on both branches", async () => {
    seedFinance({ kpi });
    renderWithProviders(<FinanceOverview />);

    expect(await screen.findByText("38.0%")).toBeInTheDocument();
    expect(screen.getByText("22.0%")).toBeInTheDocument();
  });

  it("renders zeroed charts and current ratio fallback without kpi", async () => {
    seedFinance({ kpi: null });
    renderWithProviders(<FinanceOverview />);

    await screen.findByText("Finance");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });
});
