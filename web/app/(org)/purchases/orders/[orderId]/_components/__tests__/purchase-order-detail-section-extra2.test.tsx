import { screen, waitFor, within } from "@testing-library/react";
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

function seedPurchaseOrder(orderOverrides = {}, lists = {}) {
  const order = { ...baseOrder, ...orderOverrides };
  const { shipments = [], journals: journalList = journals } = lists as {
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
      HttpResponse.json({ success: true, message: "OK.", data: { journals: journalList } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments } }),
    ),
  );
}

async function selectJournal(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("combobox", { name: "Jurnal" }));
  await user.click(await screen.findByRole("option", { name: "Kas — cash" }));
}

beforeEach(() => {});

describe("PurchaseOrderDetail extra2", () => {
  it("cancels a draft order from the overview actions", async () => {
    seedPurchaseOrder({ state: "draft" });
    let cancelCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/cancel", () => {
        cancelCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    await waitFor(() => expect(cancelCalls).toBe(1));
  });

  it("submits the receive dialog against the receive endpoint", async () => {
    seedPurchaseOrder({ state: "confirmed" });
    let receiveCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/receive", () => {
        receiveCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));
    await user.click(await screen.findByRole("button", { name: "Terima" }));
    const dialog = await screen.findByRole("dialog");
    await selectJournal(user);
    await user.click(within(dialog).getByRole("button", { name: "Terima" }));

    await waitFor(() => expect(receiveCalls).toBe(1));
  });

  it("submits the supplier-bill dialog with the override flag", async () => {
    seedPurchaseOrder({ state: "confirmed" });
    const billBodies: unknown[] = [];
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/purchase-orders/:id/supplier-bill",
        async ({ request }) => {
          billBodies.push(await request.json());
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penagihan" }));
    await user.click(await screen.findByRole("button", { name: "Buat tagihan pemasok" }));
    const dialog = await screen.findByRole("dialog");
    await selectJournal(user);
    await user.click(within(dialog).getByRole("checkbox"));
    await user.click(within(dialog).getByRole("button", { name: "Buat tagihan pemasok" }));

    await waitFor(() => expect(billBodies).toHaveLength(1));
    expect(billBodies[0]).toMatchObject({ override: true });
  });

  it("submits the payment dialog for a confirmed order", async () => {
    seedPurchaseOrder({ state: "confirmed" });
    let payCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders/:id/pay", () => {
        payCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));
    const payButton = await screen.findByRole("button", { name: "Buat pembayaran" });
    expect(payButton).toBeEnabled();
    await user.click(payButton);
    const dialog = await screen.findByRole("dialog");
    await selectJournal(user);
    await user.click(within(dialog).getByRole("button", { name: "Buat pembayaran" }));

    await waitFor(() => expect(payCalls).toBe(1));
  });

  it("shows the empty-lines state when the order has no lines", async () => {
    seedPurchaseOrder({ state: "draft", lines: [] });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
  });

  it("hints that receiving needs confirmation for draft orders", async () => {
    seedPurchaseOrder({ state: "draft" });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Penerimaan" }));

    expect(
      await screen.findByText("Belum ada pengiriman — konfirmasi pesanan untuk membuatnya."),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Konfirmasi pesanan untuk mengaktifkan penerimaan."),
    ).toBeInTheDocument();
  });

  it("falls back to the supplier id when the supplier name is unknown", async () => {
    seedPurchaseOrder({ state: "draft" });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.getByText("#7")).toBeInTheDocument();
  });

  it("shows no supplier reference row when vendorRef is absent", async () => {
    seedPurchaseOrder({ state: "draft", vendor_ref: null });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.queryByText("VEND-99")).not.toBeInTheDocument();
  });
});
