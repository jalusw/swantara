import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { StockOverviewSection } from "../stock-overview-section";

const STAMP = "2026-01-01T00:00:00Z";

const WAREHOUSES = [{ id: 1, organization_id: 1, name: "Main Warehouse", code: "WH-01" }];

const LOCATIONS = [
  {
    id: 9,
    warehouse_id: 1,
    name: "WH/Stock",
    code: null,
    parent_id: null,
    usage: "internal",
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 10,
    warehouse_id: null,
    name: "WH/Output",
    code: null,
    parent_id: null,
    usage: "internal",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const PRODUCTS = [
  {
    id: 5,
    organization_id: 1,
    name: "Finished Widget",
    category_id: null,
    type: "stockable",
    unit_id: null,
    purchase_unit_id: null,
    list_price: 100,
    standard_cost: 60,
    is_purchasable: true,
    is_sellable: true,
    is_manufactured: false,
    tracking: "none",
    weight: 0,
    volume: 0,
    hs_code: null,
    description_sale: null,
    description_purchase: null,
    active: true,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const QUANTS = [
  { id: 1, item_id: 5, location_id: 9, batch_id: null, quantity: 100, reserved_qty: 20 },
  { id: 2, item_id: 42, location_id: 10, batch_id: null, quantity: 7, reserved_qty: 0 },
];

const MOVES = [
  {
    id: 1,
    organization_id: 1,
    shipment_id: null,
    item_id: 5,
    qty: 10,
    unit_id: null,
    src_location_id: 9,
    dst_location_id: 77,
    batch_id: null,
    state: "done",
    unit_cost: null,
    origin_type: null,
    origin_id: null,
    scheduled_date: null,
    date_done: "2026-02-05T00:00:00Z",
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    organization_id: 1,
    shipment_id: null,
    item_id: 43,
    qty: 3,
    unit_id: null,
    src_location_id: 78,
    dst_location_id: 10,
    batch_id: null,
    state: "cancelled",
    unit_cost: null,
    origin_type: null,
    origin_id: null,
    scheduled_date: null,
    date_done: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 3,
    organization_id: 1,
    shipment_id: null,
    item_id: 5,
    qty: 1,
    unit_id: null,
    src_location_id: 9,
    dst_location_id: 10,
    batch_id: null,
    state: "draft",
    unit_cost: null,
    origin_type: null,
    origin_id: null,
    scheduled_date: null,
    date_done: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 4,
    organization_id: 1,
    shipment_id: null,
    item_id: 5,
    qty: 2,
    unit_id: null,
    src_location_id: 9,
    dst_location_id: 10,
    batch_id: null,
    state: "confirmed",
    unit_cost: null,
    origin_type: null,
    origin_id: null,
    scheduled_date: null,
    date_done: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function useHandlers(options?: { failBalances?: boolean; emptyMoves?: boolean }) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: WAREHOUSES } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-locations", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { locations: LOCATIONS } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock/balances", () => {
      if (options?.failBalances) {
        return HttpResponse.json({ success: false, message: "Balances down." }, { status: 500 });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { balances: QUANTS } });
    }),
    http.get("*/api/v1/organizations/:organizationId/stock-movements", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { movements: options?.emptyMoves ? [] : MOVES },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: PRODUCTS } }),
    ),
  );
}

beforeEach(() => {});

describe("StockOverviewSection branches", () => {
  it("aggregates balances with item fallbacks", async () => {
    useHandlers();
    renderWithProviders(<StockOverviewSection orgId="1" />);
    expect(await screen.findByText("Finished Widget")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByText("WH/Stock")).toBeInTheDocument();
  });

  it("filters on-hand rows by location", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<StockOverviewSection orgId="1" />);
    await screen.findByText("Finished Widget");
    await user.click(screen.getByRole("combobox", { name: "Filter by location" }));
    await user.click(await screen.findByRole("option", { name: "WH/Output" }));
    await waitFor(() => expect(screen.queryByText("WH/Stock")).not.toBeInTheDocument());
    expect(screen.getAllByText("WH/Output").length).toBeGreaterThanOrEqual(2);
  });

  it("renders move states with fallbacks", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<StockOverviewSection orgId="1" />);
    await screen.findByText("Finished Widget");
    await user.click(screen.getByRole("tab", { name: "Stock movements" }));
    expect(await screen.findByText("Stock movement ledger")).toBeInTheDocument();
    expect(screen.getByText("done")).toBeInTheDocument();
    expect(screen.getByText("cancelled")).toBeInTheDocument();
    expect(screen.getByText("draft")).toBeInTheDocument();
    expect(screen.getByText("confirmed")).toBeInTheDocument();
    expect(screen.getByText("77")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the moves empty state", async () => {
    const user = userEvent.setup();
    useHandlers({ emptyMoves: true });
    renderWithProviders(<StockOverviewSection orgId="1" />);
    await screen.findByText("Finished Widget");
    await user.click(screen.getByRole("tab", { name: "Stock movements" }));
    expect(await screen.findByText("No stock movements found.")).toBeInTheDocument();
  });

  it("shows the error state with retry when balances fail", async () => {
    useHandlers({ failBalances: true });
    renderWithProviders(<StockOverviewSection orgId="1" />);
    expect(await screen.findByText("Balances down.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });
});
