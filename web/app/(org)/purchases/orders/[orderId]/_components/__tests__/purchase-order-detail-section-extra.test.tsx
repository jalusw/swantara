import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseOrderDetail } from "../purchase-order-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const baseOrder = {
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

function seedPurchaseOrder(orderOverrides = {}, lists = {}) {
  const order = { ...baseOrder, ...orderOverrides };
  const { shipments = [], journals = [] } = lists as {
    shipments?: unknown[];
    journals?: unknown[];
  };
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
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
  );
}

beforeEach(() => {});

describe("PurchaseOrderDetail extra", () => {
  it("confirms a draft order from the overview actions", async () => {
    seedPurchaseOrder({ state: "draft" });
    let confirmCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/confirm", () => {
        confirmCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(confirmCalls).toBe(1));
  });

  it("shows the edit-disabled hint once confirmed", async () => {
    seedPurchaseOrder({ state: "confirmed" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.getByText("Hanya pesanan berstatus draf yang dapat diubah.")).toBeInTheDocument();
  });

  it("shows supplier reference and expected date when present", async () => {
    seedPurchaseOrder({ vendor_ref: "VEND-99", expected_date: "2026-04-01" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.getByText("VEND-99")).toBeInTheDocument();
  });

  it("shows the not-found state when the order is missing", async () => {
    seedPurchaseOrder();
    server.use(
      http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect(await screen.findByText("Pesanan pembelian tidak ditemukan.")).toBeInTheDocument();
  });

  it("shows the linked shipment on the receipt tab", async () => {
    seedPurchaseOrder(
      { state: "confirmed" },
      { shipments: [{ id: 8, origin: "PO-0005", name: "IN-01", state: "assigned" }] },
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));
    expect(await screen.findByText(/IN-01/)).toBeInTheDocument();
  });

  it("lists billable lines and opens the supplier-bill dialog", async () => {
    seedPurchaseOrder({ state: "confirmed" });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(await screen.findByText(/Baris yang akan ditagih/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Buat tagihan pemasok" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("disables billing when every line is fully billed", async () => {
    seedPurchaseOrder({
      state: "confirmed",
      lines: [{ ...baseOrder.lines[0], qty_billed: 2 }],
    });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));

    expect(
      await screen.findByText("Tidak ada baris yang sudah diterima tetapi belum ditagih."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Buat tagihan pemasok" })).toBeDisabled();
  });

  it("disables payment for draft orders", async () => {
    seedPurchaseOrder({ state: "draft" });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));

    expect(await screen.findByRole("button", { name: "Buat pembayaran" })).toBeDisabled();
  });
});
