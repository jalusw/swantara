import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ManufacturingSection } from "../production-section";

const kpi = {
  oee_pct: 0.85,
  yield_pct: 0.96,
  scrap_pct: 0.02,
  cost_variance_pct: 0.015,
  order_count: 24,
};

function seedManufacturing(payload: unknown) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/kpis/manufacturing", () =>
      HttpResponse.json({ success: true, message: "OK.", data: payload }),
    ),
  );
}

beforeEach(() => {});

describe("ManufacturingSection", () => {
  it("renders order count when manufacturing kpi is present", async () => {
    seedManufacturing({ kpi });
    renderWithProviders(<ManufacturingSection />);

    expect(await screen.findByText("Manufaktur")).toBeInTheDocument();
    expect((await screen.findAllByText("24")).length).toBeGreaterThan(0);
  });

  it("renders placeholders when manufacturing kpi is missing", async () => {
    seedManufacturing({});
    renderWithProviders(<ManufacturingSection />);

    await screen.findByText("Manufaktur");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders percentage branches from kpi values", async () => {
    seedManufacturing({ kpi });
    renderWithProviders(<ManufacturingSection />);

    expect((await screen.findAllByText("85.0%")).length).toBeGreaterThan(0);
    expect((await screen.findAllByText("96.0%")).length).toBeGreaterThan(0);
  });

  it("falls back to zeroed datasets with null kpi", async () => {
    seedManufacturing({ kpi: null });
    renderWithProviders(<ManufacturingSection />);

    await screen.findByText("Manufaktur");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });
});
