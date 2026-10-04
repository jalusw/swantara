import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseOrderDetail } from "../purchase-order-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const VENDOR = {
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

const WAREHOUSE = {
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

const JOURNAL = {
  id: 3,
  organization_id: 1,
  name: "Purchase Journal",
  code: "PUJ",
  type: "purchase",
  default_account_id: null,
  bank_account_id: null,
  created_at: STAMP,
  updated_at: STAMP,
};

function makeOrder(over: Record<string, unknown> = {}) {
  return {
    id: 5,
    organization_id: 1,
    name: "PO-0005",
    supplier_id: 7,
    vendor_ref: "VREF-1",
    currency_code: null,
    warehouse_id: 1,
    dest_location_id: null,
    state: "confirmed",
    order_date: "2026-02-01",
    expected_date: "2026-02-10",
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
    ...over,
  };
}

let currentOrder: ReturnType<typeof makeOrder> = makeOrder();
let confirmCalled = false;
let cancelCalled = false;

function useHandlers() {
  confirmCalled = false;
  cancelCalled = false;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order: currentOrder } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [VENDOR] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [WAREHOUSE] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [JOURNAL] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          shipments: [{ id: 9, name: "PK-0009", origin: "PO-0005", state: "done" }],
        },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/confirm", () => {
      confirmCalled = true;
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/cancel", () => {
      cancelCalled = true;
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/receive", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/supplier-bill", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/pay", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {
  currentOrder = makeOrder();
  useHandlers();
});

describe("PurchaseOrderDetail branches3", () => {
  it("renders confirmed order meta with supplier ref and linked shipment", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect((await screen.findAllByText("PO-0005")).length).toBeGreaterThan(0);
    expect(screen.getByText("VREF-1")).toBeInTheDocument();
    expect(screen.getByText("Main Warehouse")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));
    expect(await screen.findByText("PK-0009 — done")).toBeInTheDocument();
  });

  it("confirms a draft order from the overview action", async () => {
    const user = userEvent.setup();
    currentOrder = makeOrder({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(confirmCalled).toBe(true));
  });

  it("cancels a draft order from the overview action", async () => {
    const user = userEvent.setup();
    currentOrder = makeOrder({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    await waitFor(() => expect(cancelCalled).toBe(true));
  });

  it("shows the empty-lines state when the order has no lines", async () => {
    currentOrder = makeOrder({ lines: [] });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect((await screen.findAllByText("PO-0005")).length).toBeGreaterThan(0);
    expect(screen.getByText("Belum ada baris")).toBeInTheDocument();
  });

  it("opens the bill dialog when partially-billed lines exist", async () => {
    const user = userEvent.setup();
    currentOrder = makeOrder({
      lines: [
        {
          id: 21,
          order_id: 5,
          sequence: 10,
          item_id: 5,
          description: null,
          qty_ordered: 2,
          qty_received: 2,
          qty_billed: 1,
          qty_returns: 0,
          unit_id: null,
          unit_price: 50,
          discount_pct: 0,
          tax_ids: [],
          dimension_id: null,
          price_subtotal: 100,
        },
      ],
    });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));
    await user.click(screen.getByRole("button", { name: "Buat tagihan pemasok" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("renders the payments tab with a disabled pay action for drafts", async () => {
    const user = userEvent.setup();
    currentOrder = makeOrder({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));
    expect(await screen.findByRole("button", { name: "Buat pembayaran" })).toBeDisabled();
  });

  it("renders missing warehouse and dates with fallbacks", async () => {
    currentOrder = makeOrder({ warehouse_id: null, order_date: null, expected_date: null });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect((await screen.findAllByText("PO-0005")).length).toBeGreaterThan(0);
    expect(screen.queryByText("Main Warehouse")).toBeNull();
  });
});
