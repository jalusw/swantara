import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseOrdersSection } from "../purchase-orders-section";

const STAMP = "2026-01-01T00:00:00Z";

const supplier = {
  id: 7,
  organization_id: 1,
  name: "Acme Supplier Co",
  display_name: "Acme Supplier",
  is_organization: true,
  parent_id: null,
  email: null,
  phone: null,
  mobile: null,
  website: null,
  tax_id: null,
  industry: null,
  currency_code: "USD",
  lang: "en",
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const order = {
  id: 1,
  organization_id: 1,
  name: "PO-0001",
  supplier_id: 7,
  vendor_ref: null,
  currency_code: null,
  warehouse_id: 1,
  dest_location_id: null,
  state: "draft",
  order_date: "2026-02-01",
  expected_date: null,
  payment_term_id: null,
  incoterm: null,
  amount_untaxed: 100,
  amount_tax: 10,
  amount_total: 110,
  invoice_status: "no",
  receipt_status: "pending",
  created_at: STAMP,
  updated_at: STAMP,
};

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

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-orders", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { orders: [order] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [warehouse] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
});

describe("PurchaseOrdersSection", () => {
  it("renders seeded orders with supplier names", async () => {
    renderWithProviders(<PurchaseOrdersSection orgId="1" />);

    expect(await screen.findByText("PO-0001")).toBeInTheDocument();
    expect(screen.getByText("Acme Supplier")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrdersSection orgId="1" />);

    await screen.findByText("PO-0001");
    await user.click(screen.getByRole("button", { name: "New purchase order" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "New purchase order" })).toBeInTheDocument();
  });
});
