import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AccountingStats } from "../accounting-stats";

function useLocalKpis() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/kpis/finance", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          kpi: {
            revenue: 84320,
            expenses: 52300,
            gross_margin_pct: 0.38,
            net_margin_pct: 0.22,
            ebitda: 32020,
            current_ratio: 1.8,
          },
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/kpis/cash", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { kpi: { position: 125000, burn: 15000, forecast: 140000 } },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalKpis();
});

describe("AccountingStats", () => {
  it("renders revenue, expenses, profit and cash labels", async () => {
    renderWithProviders(<AccountingStats />);

    expect(await screen.findByText("Pendapatan")).toBeInTheDocument();
    expect(screen.getByText("Biaya")).toBeInTheDocument();
    expect(screen.getByText("Laba bersih")).toBeInTheDocument();
    expect(screen.getByText("Saldo kas")).toBeInTheDocument();
  });

  it("renders seeded kpi values", async () => {
    renderWithProviders(<AccountingStats />);

    expect(await screen.findByText("Rp 84.320,00")).toBeInTheDocument();
    expect(screen.getByText("Rp 52.300,00")).toBeInTheDocument();
    expect(screen.getByText("Rp 32.020,00")).toBeInTheDocument();
    expect(screen.getByText("Rp 125.000,00")).toBeInTheDocument();
  });

  it("renders trend hints from kpi margins", async () => {
    renderWithProviders(<AccountingStats />);

    expect(await screen.findByText("Margin bersih: 22%")).toBeInTheDocument();
    expect(screen.getByText("Prakiraan: Rp 140.000,00")).toBeInTheDocument();
  });
});
