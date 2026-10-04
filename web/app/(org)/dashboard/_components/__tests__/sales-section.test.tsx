import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PipelineSection, SalesSection } from "../sales-section";

beforeEach(() => {
  server.use(
    http.get("*/kpis/sales", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          kpi: {
            bookings: 1000,
            revenue: 800,
            cogs: 300,
            gross_margin: 500,
            gross_margin_pct: 0.625,
            win_rate: 0.6,
          },
        },
      }),
    ),
    http.get("*/kpis/pipeline", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          kpi: {
            total_expected_revenue: 2000,
            weighted_pipeline: 1500,
            win_rate: 0.5,
            stages: 4,
          },
        },
      }),
    ),
  );
});

describe("SalesSection", () => {
  it("renders seeded sales KPIs", async () => {
    renderWithProviders(<SalesSection />);

    expect((await screen.findAllByText("60.0%")).length).toBeGreaterThan(0);
    expect((await screen.findAllByText("Bookings")).length).toBeGreaterThan(0);
  });

  it("keeps sales content visible when clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SalesSection />);

    await screen.findAllByText("60.0%");
    await user.click(screen.getByText("Penjualan"));

    expect((await screen.findAllByText("60.0%")).length).toBeGreaterThan(0);
  });
});

describe("PipelineSection", () => {
  it("renders seeded pipeline KPIs", async () => {
    renderWithProviders(<PipelineSection />);

    expect((await screen.findAllByText("50.0%")).length).toBeGreaterThan(0);
    expect((await screen.findAllByText("Pipa penjualan tertimbang")).length).toBeGreaterThan(0);
  });
});
