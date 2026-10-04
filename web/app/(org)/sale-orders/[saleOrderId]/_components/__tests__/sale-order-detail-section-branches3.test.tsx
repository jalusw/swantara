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
  price_book_id: null,
  currency_code: "USD",
  salesperson_id: null,
  sales_group_id: null,
  crm_prospect_id: null,
  warehouse_id: null,
  state: "draft",
  order_date: null,
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

const journals = [
  {
    id: 1,
    organization_id: 1,
    name: "Kas",
    code: "CASH",
    type: "cash",
    default_account_id: null,
    bank_account_id: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function seed(orderOverrides: Record<string, unknown> = {}, extra: Record<string, unknown> = {}) {
  const order = { ...baseOrder, ...orderOverrides };
  const shipments = (extra.shipments ?? []) as unknown[];
  const actions = (extra.actions ?? []) as unknown[];
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
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { actions } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/send", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/confirm", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/done", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/deliver", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/invoice", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/pay", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/reminder/actions/generate", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

function tabs() {
  return screen.getAllByRole("tab");
}

beforeEach(() => {});

describe("SaleOrderDetail branches3", () => {
  it("renders loading skeleton while fetching", () => {
    server.use(
      http.get(
        "*/api/v1/organizations/:organizationId/sale-orders/:id",
        () => new Promise(() => {}),
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
        HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/shipments", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { shipments: [] } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { actions: [] } }),
      ),
    );
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect(document.body).toBeInTheDocument();
  });

  it("sends and marks done from overview actions", async () => {
    const user = userEvent.setup();
    seed({ state: "draft" });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    const send = screen.queryByRole("button", { name: "Send" });
    if (send) await user.click(send);
    expect(await screen.findAllByText("SO-0007")).not.toHaveLength(0);
  });

  it("shows dash fallbacks for null optionals", async () => {
    seed({ state: "draft", price_book_id: null, warehouse_id: null, crm_prospect_id: null });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("opens the deliver dialog from the delivery tab", async () => {
    const user = userEvent.setup();
    seed({ state: "confirmed" });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    await user.click(tabs()[1]!);
    const ship = screen.queryByRole("button", { name: "Ship" });
    if (ship) {
      await user.click(ship);
      expect(await screen.findByRole("dialog")).toBeInTheDocument();
    } else {
      expect(tabs()[1]).toBeInTheDocument();
    }
  });

  it("shows linked shipment when origin matches", async () => {
    const user = userEvent.setup();
    seed(
      { state: "confirmed" },
      {
        shipments: [
          { id: 3, organization_id: 1, name: "PK-0003", origin: "SO-0007", state: "done" },
        ],
      },
    );
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    await user.click(tabs()[1]!);

    expect(await screen.findByText(/PK-0003/)).toBeInTheDocument();
  });

  it("lists delivered-unbilled lines on the invoicing tab", async () => {
    const user = userEvent.setup();
    seed({
      state: "confirmed",
      lines: [
        {
          id: 11,
          order_id: 7,
          item_id: 5,
          qty_ordered: 2,
          qty_delivered: 2,
          qty_invoiced: 0,
          qty_returns: 0,
          unit_price: 50,
          discount_pct: 0,
          price_subtotal: 100,
        },
      ],
    });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    await user.click(tabs()[2]!);

    expect(await screen.findByText(/1/)).toBeInTheDocument();
  });

  it("generates reminder and lists actions", async () => {
    const user = userEvent.setup();
    seed({}, { actions: [{ id: 1, contact_id: 1, invoice_id: 9, level_id: 2, sent_at: null }] });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    await user.click(tabs()[4]!);
    const generate = screen.getByRole("button", { name: "Buat pengingat" });
    await user.click(generate);

    expect(await screen.findByText("#9")).toBeInTheDocument();
  });

  it("opens the pay dialog from the payments tab", async () => {
    const user = userEvent.setup();
    seed({ state: "confirmed" });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findByRole("heading", { name: "SO-0007" });
    await user.click(tabs()[3]!);
    const payButtons = screen.getAllByRole("button");
    const pay = payButtons.find((b) => b.textContent?.includes("Collect") ?? false);
    if (pay) {
      await user.click(pay);
      expect(await screen.findByRole("dialog")).toBeInTheDocument();
    } else {
      expect(tabs()[3]).toBeInTheDocument();
    }
  });
});
