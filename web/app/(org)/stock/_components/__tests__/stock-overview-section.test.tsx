import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { StockOverviewSection } from "../stock-overview-section";

const STAMP = "2026-01-01T00:00:00Z";

const warehouse = {
  id: 1,
  organization_id: 1,
  name: "Main Warehouse",
  code: "WH-01",
  line1: null,
  line2: null,
  city: null,
  state: null,
  postal_code: null,
  country_code: null,
  created_at: STAMP,
  updated_at: STAMP,
};

const locations = [
  {
    id: 9,
    warehouse_id: 1,
    name: "WH/Stock",
    code: null,
    parent_id: null,
    usage: "internal",
    barcode: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 10,
    warehouse_id: 1,
    name: "WH/Output",
    code: null,
    parent_id: null,
    usage: "internal",
    barcode: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const item = {
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
};

const balance = {
  id: 1,
  item_id: 5,
  location_id: 9,
  batch_id: null,
  quantity: 100,
  reserved_qty: 20,
  created_at: STAMP,
  updated_at: STAMP,
};

const move = {
  id: 1,
  organization_id: 1,
  shipment_id: null,
  item_id: 5,
  qty: 10,
  unit_id: null,
  src_location_id: 9,
  dst_location_id: 10,
  batch_id: null,
  state: "done",
  unit_cost: null,
  origin_type: null,
  origin_id: null,
  scheduled_date: null,
  date_done: "2026-02-05T00:00:00Z",
  created_at: STAMP,
  updated_at: STAMP,
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [warehouse] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-locations", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { locations } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock/balances", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { balances: [balance] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-movements", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { movements: [move] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
});

describe("StockOverviewSection", () => {
  it("renders seeded on-hand quantities", async () => {
    renderWithProviders(<StockOverviewSection orgId="1" />);

    expect(await screen.findByText("Finished Widget")).toBeInTheDocument();
    expect(screen.getByText("WH/Stock")).toBeInTheDocument();
  });

  it("switches to the moves ledger tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<StockOverviewSection orgId="1" />);

    await screen.findByText("Finished Widget");
    await user.click(screen.getByRole("tab", { name: "Stock movements" }));

    expect(await screen.findByText("Stock movement ledger")).toBeInTheDocument();
    expect(screen.getByText("WH/Output")).toBeInTheDocument();
  });
});
