import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SaleOrderDetail } from "../sale-order-detail-section";

const order = {
  id: 7,
  organization_id: 1,
  name: "SO-0007",
  contact_id: 1,
  ship_address_id: null,
  bill_address_id: null,
  price_book_id: 1,
  currency_code: "USD",
  salesperson_id: null,
  sales_group_id: null,
  crm_prospect_id: null,
  warehouse_id: 1,
  state: "confirmed",
  order_date: "2026-03-01",
  expected_date: null,
  validity_date: null,
  payment_term_id: null,
  incoterm: null,
  customer_po_ref: null,
  amount_untaxed: 100,
  amount_tax: 10,
  amount_total: 110,
  invoice_status: "to_invoice",
  delivery_status: "partial",
  note: null,
  lines: [
    {
      id: 11,
      order_id: 7,
      sequence: 10,
      item_id: 5,
      description: null,
      qty_ordered: 2,
      qty_delivered: 1,
      qty_invoiced: 0,
      qty_returns: 0,
      unit_id: null,
      unit_price: 50,
      discount_pct: 0,
      tax_ids: [],
      dimension_id: null,
      price_subtotal: 100,
      price_tax: 10,
      price_total: 110,
    },
  ],
  created_at: "2026-03-01T00:00:00Z",
  updated_at: "2026-03-01T00:00:00Z",
};

function useDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/sale-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          contacts: [
            {
              id: 1,
              organization_id: 1,
              name: "Bluebird Trading Pte. Ltd.",
              display_name: "Bluebird Trading",
              is_organization: true,
              email: "billing@bluebird.sg",
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          price_books: [
            {
              id: 1,
              name: "Retail",
              currency_code: "USD",
              organization_id: 1,
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          warehouses: [
            {
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
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { actions: [] } }),
    ),
  );
}

beforeEach(() => {
  useDetailHandlers();
});

describe("SaleOrderDetail", () => {
  it("renders the order header with customer and totals", async () => {
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect((await screen.findAllByText("SO-0007")).length).toBeGreaterThan(0);
    expect(screen.getByText("Bluebird Trading")).toBeInTheDocument();
    expect(screen.getByText("Order header")).toBeInTheDocument();
    expect(screen.getByText("Retail")).toBeInTheDocument();
    expect(screen.getByText("Main Warehouse")).toBeInTheDocument();
  });

  it("switches to the delivery tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Delivery" }));

    expect(await screen.findByText("Ship / Deliver")).toBeInTheDocument();
  });

  it("switches to the invoicing tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Invoicing" }));

    expect(await screen.findByText("Create invoice")).toBeInTheDocument();
  });
});
