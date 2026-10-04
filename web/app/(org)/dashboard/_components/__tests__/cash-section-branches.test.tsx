import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CashSection } from "../cash-section";

const CASH = { position: 125000, burn: 15000, forecast: 140000 };
const AR_AP = { dso: 32.5, dpo: 28.3, overdue_ar_pct: 0.12, overdue_ap_pct: 0.08 };

function useCashHandlers(cash: unknown, arAp: unknown) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/kpis/cash", () =>
      cash instanceof HttpResponse
        ? cash
        : HttpResponse.json({ success: true, message: "OK.", data: { kpi: cash } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/kpis/ar-ap", () =>
      arAp instanceof HttpResponse
        ? arAp
        : HttpResponse.json({ success: true, message: "OK.", data: { kpi: arAp } }),
    ),
  );
}

beforeEach(() => {});

describe("CashSection branches", () => {
  it("renders seeded cash and ar-ap values", async () => {
    useCashHandlers(CASH, AR_AP);
    renderWithProviders(<CashSection />);
    expect(await screen.findByText("Rp 125.000,00")).toBeInTheDocument();
    expect(screen.getByText("Rp 15.000,00")).toBeInTheDocument();
    expect(screen.getByText("Rp 140.000,00")).toBeInTheDocument();
    expect(screen.getByText("32.5d")).toBeInTheDocument();
    expect(screen.getByText("28.3d")).toBeInTheDocument();
    expect(screen.getByText("12.0% piutang tertunggak")).toBeInTheDocument();
    expect(screen.getByText("8.0% utang tertunggak")).toBeInTheDocument();
  });

  it("renders chart regions with formatted values", async () => {
    useCashHandlers(CASH, AR_AP);
    renderWithProviders(<CashSection />);
    await screen.findByText("Rp 125.000,00");
    expect(screen.getByRole("img", { name: /Kas/ })).toBeInTheDocument();
    expect(screen.getByRole("img", { name: /^Piutang & Utang/ })).toBeInTheDocument();
  });

  it("falls back to placeholders when cash kpi is missing", async () => {
    useCashHandlers(null, AR_AP);
    renderWithProviders(<CashSection />);
    expect(await screen.findByText("32.5d")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("falls back to placeholders when ar-ap kpi is missing", async () => {
    useCashHandlers(CASH, null);
    renderWithProviders(<CashSection />);
    expect(await screen.findByText("Rp 125.000,00")).toBeInTheDocument();
    expect((await screen.findAllByText("—")).length).toBeGreaterThan(0);
  });

  it("renders placeholders when both kpis fail", async () => {
    useCashHandlers(
      HttpResponse.json({ success: false, message: "Down." }, { status: 500 }),
      HttpResponse.json({ success: false, message: "Down." }, { status: 500 }),
    );
    renderWithProviders(<CashSection />);
    expect((await screen.findAllByText("—")).length).toBeGreaterThan(0);
  });
});
