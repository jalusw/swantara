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

function useSaleOrderHandlers(
  order: Record<string, unknown>,
  opts?: { shipments?: unknown[]; reminder?: unknown[] },
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
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { shipments: opts?.shipments ?? [] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/reminder/actions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { actions: opts?.reminder ?? [] } }),
    ),
  );
}

beforeEach(() => {});

describe("SaleOrderDetail remainder", () => {
  it("shows the not-found state when the order is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/sale-orders/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="999" />);

    expect(await screen.findByText("Pesanan tidak ditemukan.")).toBeInTheDocument();
  });

  it("sends a draft order", async () => {
    let sent = false;
    useSaleOrderHandlers(orderFixture());
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/send", () => {
        sent = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Kirim" }));

    await waitFor(() => expect(sent).toBe(true));
  });

  it("confirms a sent order", async () => {
    let confirmed = false;
    useSaleOrderHandlers(orderFixture({ state: "sent" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/confirm", () => {
        confirmed = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(confirmed).toBe(true));
  });

  it("cancels and marks done from the overview actions", async () => {
    let cancelled = false;
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/cancel", () => {
        cancelled = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    await waitFor(() => expect(cancelled).toBe(true));
  });

  it("shows the edit-disabled hint for confirmed orders", async () => {
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect(await screen.findByText("Hanya pesanan draf yang bisa diubah.")).toBeInTheDocument();
  });

  it("shows the linked shipment on the delivery tab", async () => {
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }), {
      shipments: [{ id: 3, origin: "SO-0007", name: "PK-0003", state: "assigned" }],
    });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengiriman" }));

    expect(await screen.findByText(/PK-0003/)).toBeInTheDocument();
  });

  it("shows the no-shipment message for draft orders", async () => {
    useSaleOrderHandlers(orderFixture());
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengiriman" }));

    expect(
      await screen.findByText("Belum ada pengiriman — konfirmasi pesanan untuk membuatnya."),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Konfirmasi pesanan untuk mengaktifkan pengiriman."),
    ).toBeInTheDocument();
  });

  it("opens the deliver dialog for confirmed orders", async () => {
    let delivered: unknown = null;
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/sale-orders/:id/deliver",
        async ({ request }) => {
          delivered = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: {
            journals: [
              {
                id: 2,
                organization_id: 1,
                name: "Delivery Journal",
                code: "DLV",
                type: "general",
                default_account_id: null,
                bank_account_id: null,
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            ],
          },
        }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengiriman" }));
    await user.click(await screen.findByRole("button", { name: "Kirim / Antar" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByLabelText("Jurnal"));
    await user.click(
      await screen.findByRole("option", { name: "Delivery Journal \u2014 general" }),
    );
    await user.click(within(dialog).getByRole("button", { name: "Kirim / Antar" }));

    await waitFor(() => expect(delivered).toMatchObject({ journal_id: 2 }));
  });

  it("collects a payment from the payments tab", async () => {
    let paid: unknown = null;
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/sale-orders/:id/pay",
        async ({ request }) => {
          paid = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: {
            journals: [
              {
                id: 4,
                organization_id: 1,
                name: "Cash Journal",
                code: "CSH",
                type: "cash",
                default_account_id: null,
                bank_account_id: null,
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            ],
          },
        }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));
    await user.click(await screen.findByRole("button", { name: "Tarik pembayaran" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByLabelText("Jurnal"));
    await user.click(await screen.findByRole("option", { name: "Cash Journal \u2014 cash" }));
    await user.type(within(dialog).getByRole("spinbutton"), "50");
    await user.click(within(dialog).getByRole("button", { name: "Tarik pembayaran" }));

    await waitFor(() => expect(paid).toMatchObject({ journal_id: 4, amount: 50 }));
  });

  it("generates reminder actions from the reminder tab", async () => {
    let generated = false;
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/reminder/actions/generate", () => {
        generated = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengingat" }));
    await user.click(await screen.findByRole("button", { name: "Buat pengingat" }));

    await waitFor(() => expect(generated).toBe(true));
  });

  it("shows reminder rows when actions exist", async () => {
    useSaleOrderHandlers(orderFixture({ state: "confirmed" }), {
      reminder: [{ id: 1, contact_id: 1, invoice_id: 9, level_id: 2, sent_at: null }],
    });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengingat" }));

    expect(await screen.findByText("#9")).toBeInTheDocument();
  });

  it("shows the no-lines message when the order has no lines", async () => {
    useSaleOrderHandlers(orderFixture({ lines: [] }));
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
  });
});
