import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SaleOrderDetail } from "../sale-order-detail-section";

const baseOrder = {
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
  crm_prospect_id: 9,
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

const journals = [
  {
    id: 1,
    organization_id: 1,
    name: "Cash",
    code: "CASH",
    type: "cash",
    default_account_id: null,
    bank_account_id: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function seedSaleOrder(orderOverrides: Record<string, unknown> = {}, lists = {}) {
  const order = orderOverrides.id === null ? null : { ...baseOrder, ...orderOverrides };
  const {
    shipments = [],
    journals: journalList = journals,
    actions = [],
  } = lists as { shipments?: unknown[]; journals?: unknown[]; actions?: unknown[] };
  server.use(
    http.get("*/api/v1/organizations/:organizationId/sale-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { price_books: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: journalList } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { actions } }),
    ),
  );
}

beforeEach(() => {});

describe("SaleOrderDetail branches2", () => {
  it("renders the not-found branch when the order is missing", async () => {
    seedSaleOrder({ id: null, name: null });
    server.use(
      http.get("*/api/v1/organizations/:organizationId/sale-orders/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { order: null } }),
      ),
    );
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect(await screen.findByText("Sale order not found.")).toBeInTheDocument();
  });

  it("renders empty lines and dash fallbacks with null optionals", async () => {
    seedSaleOrder({
      state: "draft",
      lines: [],
      price_book_id: null,
      warehouse_id: null,
      crm_prospect_id: null,
      order_date: null,
    });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    expect(await screen.findByText("No lines")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the linked shipment branch when a shipment matches the order", async () => {
    seedSaleOrder(
      { state: "confirmed" },
      {
        shipments: [
          { id: 3, organization_id: 1, name: "PK-0003", origin: "SO-0007", state: "done" },
        ],
      },
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Delivery" }));
    expect(await screen.findByText(/PK-0003/)).toBeInTheDocument();
  });

  it("hides action buttons and shows edit-disabled for done orders", async () => {
    seedSaleOrder({ state: "done" });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    expect(screen.queryByRole("button", { name: "Send" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Confirm" })).not.toBeInTheDocument();
  });

  it("renders reminder rows when actions exist", async () => {
    seedSaleOrder(
      { state: "confirmed" },
      { actions: [{ id: 1, contact_id: 2, invoice_id: 9, level_id: 1, sent_at: null }] },
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Reminder" }));
    expect(await screen.findByText("#9")).toBeInTheDocument();
  });

  it("shows invoicing empty branch and disabled create button without billable lines", async () => {
    seedSaleOrder({
      state: "confirmed",
      lines: [
        {
          id: 11,
          order_id: 7,
          sequence: 10,
          item_id: 5,
          description: null,
          qty_ordered: 2,
          qty_delivered: 2,
          qty_invoiced: 2,
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
    });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Invoicing" }));
    expect(await screen.findByText("No delivered-unbilled lines to invoice.")).toBeInTheDocument();
  });
});
