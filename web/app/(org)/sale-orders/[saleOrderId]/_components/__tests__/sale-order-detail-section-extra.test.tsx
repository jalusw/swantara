import { screen, waitFor } from "@testing-library/react";
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

function seedSaleOrder(orderOverrides = {}, lists = {}) {
  const order = { ...baseOrder, ...orderOverrides };
  const { shipments = [], journals = [] } = lists as {
    shipments?: unknown[];
    journals?: unknown[];
  };
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
      HttpResponse.json({ success: true, message: "OK.", data: { actions: [] } }),
    ),
  );
}

beforeEach(() => {});

describe("SaleOrderDetail extra", () => {
  it("sends a draft order from the overview actions", async () => {
    seedSaleOrder({ state: "draft" });
    let sendCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/send", () => {
        sendCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Kirim" }));

    await waitFor(() => expect(sendCalls).toBe(1));
  });

  it("marks a confirmed order as done", async () => {
    seedSaleOrder({ state: "confirmed" });
    let doneCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders/:id/done", () => {
        doneCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("button", { name: "Tandai selesai" }));

    await waitFor(() => expect(doneCalls).toBe(1));
  });

  it("shows the edit-disabled hint for a done order", async () => {
    seedSaleOrder({ state: "done" });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    expect(screen.getByText("Hanya pesanan draf yang bisa diubah.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tandai selesai" })).not.toBeInTheDocument();
  });

  it("shows the empty-lines state when the order has no lines", async () => {
    seedSaleOrder({ state: "draft", lines: [] });
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
  });

  it("shows the not-found state when the order is missing", async () => {
    seedSaleOrder();
    server.use(
      http.get("*/api/v1/organizations/:organizationId/sale-orders/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    expect(await screen.findByText("Pesanan tidak ditemukan.")).toBeInTheDocument();
  });

  it("shows the linked shipment on the delivery tab", async () => {
    seedSaleOrder(
      { state: "confirmed" },
      {
        shipments: [{ id: 3, origin: "SO-0007", name: "PICK-01", state: "assigned" }],
      },
    );
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengiriman" }));

    expect(await screen.findByText(/PICK-01/)).toBeInTheDocument();
  });

  it("opens the deliver dialog from the delivery tab", async () => {
    seedSaleOrder({ state: "confirmed" });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pengiriman" }));
    await user.click(await screen.findByRole("button", { name: "Kirim / Antar" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("disables invoicing when no lines are delivered-unbilled", async () => {
    seedSaleOrder({ state: "confirmed", lines: [] });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(
      await screen.findByText("Tidak ada baris yang sudah dikirim tetapi belum ditagih."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Buat faktur" })).toBeDisabled();
  });

  it("opens the payment dialog from the payments tab", async () => {
    seedSaleOrder({ state: "confirmed" });
    const user = userEvent.setup();
    renderWithProviders(<SaleOrderDetail orgId="1" saleOrderId="7" />);

    await screen.findAllByText("SO-0007");
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));
    await user.click(await screen.findByRole("button", { name: "Tarik pembayaran" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
