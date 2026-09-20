import { screen, waitFor, within } from "@testing-library/react";
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

function seedSaleOrder(orderOverrides = {}, lists = {}) {
  const order = { ...baseOrder, ...orderOverrides };
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

async function selectJournal(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("combobox", { name: "Journal" }));
  await user.click(await screen.findByRole("option", { name: "Cash — cash" }));
}

beforeEach(() => {});

describe("SaleOrderDetail extra2", () => {
  it("confirms a sent order from the overview actions", async () => {
    seedSaleOrder({ state: "sent" });
    let confirmCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/confirm", () => {
        confirmCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Confirm" }));

    await waitFor(() => expect(confirmCalls).toBe(1));
  });

  it("cancels a draft order from the overview actions", async () => {
    seedSaleOrder({ state: "draft" });
    let cancelCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/cancel", () => {
        cancelCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(cancelCalls).toBe(1));
  });

  it("submits the deliver dialog against the deliver endpoint", async () => {
    seedSaleOrder({ state: "confirmed" });
    let deliverCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/deliver", () => {
        deliverCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Delivery" }));
    await user.click(await screen.findByRole("button", { name: "Ship / Deliver" }));
    const dialog = await screen.findByRole("dialog");
    await selectJournal(user);
    await user.click(within(dialog).getByRole("button", { name: "Ship / Deliver" }));

    await waitFor(() => expect(deliverCalls).toBe(1));
  });

  it("submits the invoice dialog against the invoice endpoint", async () => {
    seedSaleOrder({ state: "confirmed" });
    let invoiceCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/invoice", () => {
        invoiceCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Invoicing" }));
    await user.click(await screen.findByRole("button", { name: "Create invoice" }));
    const dialog = await screen.findByRole("dialog");
    await selectJournal(user);
    await user.click(within(dialog).getByRole("button", { name: "Create invoice" }));

    await waitFor(() => expect(invoiceCalls).toBe(1));
  });

  it("submits the payment dialog with an amount", async () => {
    seedSaleOrder({ state: "confirmed" });
    let payCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/pay", () => {
        payCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Payments" }));
    await user.click(await screen.findByRole("button", { name: "Collect payment" }));
    const dialog = await screen.findByRole("dialog");
    await selectJournal(user);
    await user.type(within(dialog).getByRole("spinbutton"), "50");
    await user.click(within(dialog).getByRole("button", { name: "Collect payment" }));

    await waitFor(() => expect(payCalls).toBe(1));
  });

  it("generates reminder actions from the reminder tab", async () => {
    seedSaleOrder({ state: "confirmed" });
    let generateCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/reminder/actions/generate", () => {
        generateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: { actions: [] } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Reminder" }));
    await user.click(await screen.findByRole("button", { name: "Generate reminder" }));

    await waitFor(() => expect(generateCalls).toBe(1));
  });

  it("shows the empty reminder state when no actions exist", async () => {
    seedSaleOrder({ state: "confirmed" });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Reminder" }));

    expect(await screen.findByText("No reminder actions.")).toBeInTheDocument();
  });

  it("hints that delivery needs confirmation for draft orders", async () => {
    seedSaleOrder({ state: "draft" });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Delivery" }));

    expect(
      await screen.findByText("No shipment yet — confirm the order to create one."),
    ).toBeInTheDocument();
    expect(screen.getByText("Confirm the order to enable delivery.")).toBeInTheDocument();
  });

  it("falls back to ids when contact and price_book names are unknown", async () => {
    seedSaleOrder({ state: "draft" });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    expect(screen.getAllByText("#1").length).toBeGreaterThan(0);
  });
});
