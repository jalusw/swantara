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
  lines: [],
  created_at: "2026-03-01T00:00:00Z",
  updated_at: "2026-03-01T00:00:00Z",
};

const actions = [
  {
    id: 61,
    contact_id: 1,
    invoice_id: 5,
    level_id: 2,
    sent_at: null,
  },
];

beforeEach(() => {
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
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { actions } }),
    ),
  );
});

describe("SaleOrderDetail reminder tab", () => {
  it("renders the seeded order", async () => {
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect((await screen.findAllByText("SO-0007")).length).toBeGreaterThan(0);
  });

  it("switches to the reminder tab and shows seeded actions", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Reminder" }));

    expect(await screen.findByRole("button", { name: "Generate reminder" })).toBeInTheDocument();
    expect(screen.getByText("#5")).toBeInTheDocument();
  });
});
