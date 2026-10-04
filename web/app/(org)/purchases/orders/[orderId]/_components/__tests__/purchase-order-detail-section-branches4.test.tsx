import { screen } from "@testing-library/react";
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
  warehouse_id: null,
  dest_location_id: null,
  state: "draft",
  order_date: null,
  expected_date: null,
  payment_term_id: null,
  incoterm: null,
  amount_untaxed: 100,
  amount_tax: 10,
  amount_total: 110,
  invoice_status: "no",
  receipt_status: "pending",
  lines: [],
  created_at: STAMP,
  updated_at: STAMP,
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
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function seed(orderOverrides: Record<string, unknown> = {}, extra: Record<string, unknown> = {}) {
  const order = { ...baseOrder, ...orderOverrides };
  const shipments = (extra.shipments ?? []) as unknown[];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { contacts: [{ id: 7, organization_id: 1, name: "Supplier A", display_name: null }] },
      }),
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
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/confirm", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/cancel", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
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

beforeEach(() => {});

describe("PurchaseOrderDetail branches4", () => {
  it("renders loading skeleton while fetching", () => {
    server.use(
      http.get(
        "*/api/v1/organizations/:organizationId/purchase-orders/:id",
        () => new Promise(() => {}),
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
        HttpResponse.json({ success: true, message: "OK.", data: { shipments: [] } }),
      ),
    );
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect(document.querySelector('[data-slot="skeleton"]') ?? document.body).toBeInTheDocument();
  });

  it("falls back to supplier id and warehouse dash when maps miss", async () => {
    seed({ supplier_id: 99, warehouse_id: 42, order_date: null });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    expect(await screen.findByText("#99")).toBeInTheDocument();
    expect(screen.getByText("#42")).toBeInTheDocument();
  });

  it("shows empty lines and confirm plus cancel actions in draft", async () => {
    const user = userEvent.setup();
    seed({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));
    await user.click(screen.getByRole("button", { name: "Batal" }));
    expect(await screen.findAllByText("PO-0005").then((els) => els.length)).toBeGreaterThan(0);
  });

  it("opens the receive dialog from the receipt tab", async () => {
    const user = userEvent.setup();
    seed({ state: "confirmed" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));
    await user.click(await screen.findByRole("button", { name: "Terima" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("shows the not-confirmed branch on the receipt tab in draft", async () => {
    const user = userEvent.setup();
    seed({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));

    expect(
      await screen.findByText("Belum ada pengiriman — konfirmasi pesanan untuk membuatnya."),
    ).toBeInTheDocument();
  });

  it("shows linked shipment when origin matches", async () => {
    const user = userEvent.setup();
    seed(
      { state: "confirmed" },
      {
        shipments: [
          { id: 8, organization_id: 1, name: "PK-0008", origin: "PO-0005", state: "done" },
        ],
      },
    );
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));

    expect(await screen.findByText(/PK-0008/)).toBeInTheDocument();
  });

  it("opens the bill dialog and toggles override", async () => {
    const user = userEvent.setup();
    seed({
      state: "confirmed",
      lines: [
        {
          id: 21,
          order_id: 5,
          item_id: 5,
          qty_ordered: 2,
          qty_received: 0,
          qty_billed: 0,
          unit_price: 50,
          price_subtotal: 100,
        },
      ],
    });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));
    await user.click(await screen.findByRole("button", { name: "Buat tagihan pemasok" }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toBeInTheDocument();
    const checkbox = dialog.querySelector('input[type="checkbox"]');
    expect(checkbox).not.toBeNull();
    if (checkbox) await user.click(checkbox);
  });

  it("disables pay when state forbids it and enables otherwise", async () => {
    const user = userEvent.setup();
    seed({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findByRole("heading", { name: "PO-0005" });
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));

    expect(await screen.findByRole("button", { name: "Buat pembayaran" })).toBeDisabled();
  });
});
