import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SaleOrderDetail } from "../sale-order-detail-section";

function orderFixture(overrides = {}) {
  return {
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
    state: "draft",
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
    ...overrides,
  };
}

const journal = {
  id: 2,
  organization_id: 1,
  name: "Delivery Journal",
  code: "DLV",
  type: "general",
  default_account_id: null,
  bank_account_id: null,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function useSaleOrderHandlers(
  order: Record<string, unknown>,
  opts?: { shipments?: unknown[]; journals?: unknown[] },
) {
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
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { journals: opts?.journals ?? [] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { shipments: opts?.shipments ?? [] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { actions: [] } }),
    ),
  );
}

beforeEach(() => {});

async function chooseJournal(user: ReturnType<typeof userEvent.setup>) {
  const dialog = await screen.findByRole("dialog");
  await user.click(within(dialog).getByLabelText("Journal"));
  await user.click(await screen.findByRole("option", { name: "Delivery Journal \u2014 general" }));
  return dialog;
}

describe("SaleOrderDetail branches4", () => {
  it("renders cancelled orders without workflow actions", async () => {
    useSaleOrderHandlers(orderFixture({ state: "cancelled" }));
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    expect(screen.queryByRole("button", { name: "Send" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Confirm" })).toBeNull();
  });

  it("falls back to generated names when the order name and lines are missing", async () => {
    useSaleOrderHandlers(orderFixture({ name: null, lines: null }));
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect((await screen.findAllByText("SO-7")).length).toBeGreaterThan(0);
    expect(screen.getByText("No lines")).toBeInTheDocument();
  });

  it("falls back for null line products and unnamed shipments", async () => {
    const order = orderFixture();
    const lines = (order.lines as unknown[]).map((line) => ({
      ...(line as Record<string, unknown>),
      item_id: null,
    }));
    useSaleOrderHandlers({ ...order, lines } as Record<string, unknown>, {
      shipments: [{ id: 3, origin: "SO-0007", name: null, state: "assigned" }],
    });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
    await user.click(screen.getByRole("tab", { name: "Delivery" }));

    expect(await screen.findByText(/PK-3/)).toBeInTheDocument();
  });

  it("cancels the deliver dialog", async () => {
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Delivery" }));
    await user.click(await screen.findByRole("button", { name: "Ship / Deliver" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("delivers with an empty date", async () => {
    let delivered: unknown = null;
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }), { journals: [journal] });
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/sale-orders/:id/deliver",
        async ({ request }) => {
          delivered = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Delivery" }));
    await user.click(await screen.findByRole("button", { name: "Ship / Deliver" }));

    const dialog = await chooseJournal(user);
    const dateInput = dialog.querySelector('input[type="date"]');
    if (!dateInput) throw new Error("Expected date input");
    await user.clear(dateInput);
    await user.click(within(dialog).getByRole("button", { name: "Ship / Deliver" }));

    await waitFor(() => expect(delivered).toMatchObject({ journal_id: 2, date: null }));
  });

  it("cancels the invoice dialog", async () => {
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    const tabs = screen.getAllByRole("tab");
    const invoicingTab = tabs[2];
    if (!invoicingTab) throw new Error("Expected invoicing tab");
    await user.click(invoicingTab);
    await user.click(await screen.findByRole("button", { name: "Create invoice" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("cancels the payment dialog", async () => {
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Payments" }));
    await user.click(await screen.findByRole("button", { name: "Collect payment" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
