import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseOrderDetail } from "../purchase-order-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const order = {
  id: 5,
  organization_id: 1,
  name: "PO-0005",
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
  lines: [
    {
      id: 21,
      order_id: 5,
      sequence: 10,
      item_id: 5,
      description: null,
      qty_ordered: 2,
      qty_received: 0,
      qty_billed: 0,
      qty_returns: 0,
      unit_id: null,
      unit_price: 50,
      discount_pct: 0,
      tax_ids: [],
      dimension_id: null,
      price_subtotal: 100,
    },
  ],
  created_at: STAMP,
  updated_at: STAMP,
};

const supplier = {
  id: 7,
  organization_id: 1,
  name: "Acme Supplier Co",
  display_name: "Acme Supplier",
  is_organization: true,
  email: null,
  active: true,
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

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [warehouse] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments: [] } }),
    ),
  );
});

describe("PurchaseOrderDetail", () => {
  it("renders the order header with supplier and warehouse", async () => {
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect((await screen.findAllByText("PO-0005")).length).toBeGreaterThan(0);
    expect(screen.getByText("Acme Supplier")).toBeInTheDocument();
    expect(screen.getByText("Main Warehouse")).toBeInTheDocument();
  });

  it("switches to the receipt tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Receipt" }));

    expect(
      await screen.findByText("Receive goods against this purchase order."),
    ).toBeInTheDocument();
  });
});
