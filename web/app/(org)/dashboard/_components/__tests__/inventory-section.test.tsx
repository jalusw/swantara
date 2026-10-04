import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InventoryHealthSection, InventoryOverview } from "../inventory-section";

const inventoryKpi = {
  on_hand_value: 12500.5,
  on_hand_quantity: 320,
  product_count: 48,
};

const inventoryRatioKpi = {
  turnover: 4.25,
  days_on_hand: 12.5,
  stockout_count: 3,
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/kpis/inventory", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { kpi: inventoryKpi } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/kpis/inventory-ratio", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { kpi: inventoryRatioKpi } }),
    ),
  );
});

describe("InventoryOverview", () => {
  it("renders seeded inventory KPIs", async () => {
    renderWithProviders(<InventoryOverview />);

    expect(await screen.findByText("Rp 12.500,50")).toBeInTheDocument();
    expect(screen.getAllByText("320").length).toBeGreaterThan(0);
    expect(screen.getAllByText("48").length).toBeGreaterThan(0);
  });

  it("keeps overview content visible when clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InventoryOverview />);

    await screen.findByText("Rp 12.500,50");
    await user.click(screen.getByText("Persediaan"));

    expect(screen.getByText("Rp 12.500,50")).toBeInTheDocument();
  });

  it("renders dashes when the KPI is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/kpis/inventory", () =>
        HttpResponse.json({ success: true, message: "OK.", data: {} }),
      ),
    );
    renderWithProviders(<InventoryOverview />);

    expect((await screen.findAllByText("—")).length).toBeGreaterThan(0);
    expect(screen.queryByText("Rp 12.500,50")).not.toBeInTheDocument();
  });
});

describe("InventoryHealthSection", () => {
  it("renders seeded inventory ratio KPIs", async () => {
    renderWithProviders(<InventoryHealthSection />);

    expect(await screen.findByText("4.25")).toBeInTheDocument();
    expect(screen.getByText("12.5d")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("keeps health content visible when clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InventoryHealthSection />);

    await screen.findByText("4.25");
    await user.click(screen.getAllByText("Kesehatan persediaan")[0]!);

    expect(screen.getByText("4.25")).toBeInTheDocument();
  });

  it("renders dashes when the ratio KPI is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/kpis/inventory-ratio", () =>
        HttpResponse.json({ success: true, message: "OK.", data: {} }),
      ),
    );
    renderWithProviders(<InventoryHealthSection />);

    expect((await screen.findAllByText("—")).length).toBeGreaterThan(0);
    expect(screen.queryByText("4.25")).not.toBeInTheDocument();
  });
});
