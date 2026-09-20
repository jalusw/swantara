import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseOrderDetail } from "../purchase-order-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function orderFixture(overrides = {}) {
  return {
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
    ...overrides,
  };
}

const supplier = {
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

const warehouse = {
  id: 1,
  organization_id: 1,
  name: "Main Warehouse",
  code: "WH-01",
  created_at: STAMP,
  updated_at: STAMP,
};

const journal = {
  id: 2,
  organization_id: 1,
  name: "Goods Journal",
  code: "GRN",
  type: "general",
  default_account_id: null,
  bank_account_id: null,
  created_at: STAMP,
  updated_at: STAMP,
};

function usePurchaseOrderHandlers(
  order: Record<string, unknown>,
  opts?: { shipments?: unknown[]; journals?: unknown[] },
) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { warehouses: [warehouse] },
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
  );
}

beforeEach(() => {});

describe("PurchaseOrderDetail branches5", () => {
  it("renders cancelled orders without confirm actions", async () => {
    usePurchaseOrderHandlers(orderFixture({ state: "cancelled" }));
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.queryByRole("button", { name: "Confirm" })).toBeNull();
  });

  it("falls back to generated names when the order name and lines are missing", async () => {
    usePurchaseOrderHandlers(orderFixture({ name: null, lines: null }));
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    expect((await screen.findAllByText("PO-5")).length).toBeGreaterThan(0);
    expect(screen.getByText("No lines.")).toBeInTheDocument();
  });

  it("falls back for unknown contacts, null products and unnamed shipments", async () => {
    const order = orderFixture({ supplier_id: 999 });
    const lines = (order.lines as unknown[]).map((line) => ({
      ...(line as Record<string, unknown>),
      item_id: null,
    }));
    usePurchaseOrderHandlers({ ...order, lines } as Record<string, unknown>, {
      shipments: [{ id: 9, origin: "PO-0005", name: null, state: "assigned" }],
    });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.getByText("#999")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
    await user.click(screen.getByRole("tab", { name: "Receipt" }));

    expect(await screen.findByText(/PK-9/)).toBeInTheDocument();
  });

  it("receives goods with an empty date", async () => {
    let received: unknown = null;
    usePurchaseOrderHandlers(orderFixture({ state: "confirmed" }), { journals: [journal] });
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/purchase-orders/:id/receive",
        async ({ request }) => {
          received = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Receipt" }));
    await user.click(await screen.findByRole("button", { name: "Receive" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByLabelText("Journal"));
    await user.click(await screen.findByRole("option", { name: "Goods Journal \u2014 general" }));
    const dateInput = dialog.querySelector('input[type="date"]');
    if (!dateInput) throw new Error("Expected date input");
    await user.clear(dateInput);
    await user.click(within(dialog).getByRole("button", { name: "Receive" }));

    await waitFor(() => expect(received).toMatchObject({ journal_id: 2, date: null }));
  });

  it("cancels the receive dialog", async () => {
    usePurchaseOrderHandlers(orderFixture({ state: "confirmed" }));
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Receipt" }));
    await user.click(await screen.findByRole("button", { name: "Receive" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("creates a supplier bill with override and an empty date", async () => {
    let billed: unknown = null;
    usePurchaseOrderHandlers(orderFixture({ state: "confirmed" }), { journals: [journal] });
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/purchase-orders/:id/supplier-bill",
        async ({ request }) => {
          billed = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Invoicing" }));
    await user.click(await screen.findByRole("button", { name: "Create supplier bill" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByLabelText("Journal"));
    await user.click(await screen.findByRole("option", { name: "Goods Journal \u2014 general" }));
    await user.click(within(dialog).getByRole("checkbox"));
    const dateInput = dialog.querySelector('input[type="date"]');
    if (!dateInput) throw new Error("Expected date input");
    await user.clear(dateInput);
    await user.click(within(dialog).getByRole("button", { name: "Create supplier bill" }));

    await waitFor(() =>
      expect(billed).toMatchObject({ journal_id: 2, date: null, override: true }),
    );
  });

  it("cancels the bill dialog", async () => {
    usePurchaseOrderHandlers(orderFixture({ state: "confirmed" }));
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Invoicing" }));
    await user.click(await screen.findByRole("button", { name: "Create supplier bill" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("pays with an empty date and cancels afterwards", async () => {
    let paid: unknown = null;
    usePurchaseOrderHandlers(orderFixture({ state: "confirmed" }), { journals: [journal] });
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/purchase-orders/:id/pay",
        async ({ request }) => {
          paid = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Payments" }));
    await user.click(await screen.findByRole("button", { name: "Make payment" }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByLabelText("Journal"));
    await user.click(await screen.findByRole("option", { name: "Goods Journal \u2014 general" }));
    const dateInput = dialog.querySelector('input[type="date"]');
    if (!dateInput) throw new Error("Expected date input");
    await user.clear(dateInput);
    await user.click(within(dialog).getByRole("button", { name: "Make payment" }));

    await waitFor(() => expect(paid).toMatchObject({ journal_id: 2, date: null }));
  });
});
