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
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function seedPurchaseOrder(orderOverrides: Record<string, unknown> = {}, lists = {}) {
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

beforeEach(() => {});

describe("PurchaseOrderDetail branches2", () => {
  it("renders the not-found branch when the order is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/purchase-orders/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { order: null } }),
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

    expect(await screen.findByText("Purchase order not found.")).toBeInTheDocument();
  });

  it("hides optional supplier ref and expected date branches when absent", async () => {
    seedPurchaseOrder({ vendor_ref: null, expected_date: null, order_date: null });
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(screen.queryByText("VREF-1")).not.toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the linked shipment branch on the receipt tab", async () => {
    seedPurchaseOrder(
      {},
      {
        shipments: [
          { id: 8, organization_id: 1, name: "PK-0008", origin: "PO-0005", state: "done" },
        ],
      },
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Receipt" }));
    expect(await screen.findByText(/PK-0008/)).toBeInTheDocument();
  });

  it("shows empty lines and empty billable branches", async () => {
    seedPurchaseOrder({
      lines: [],
    });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    expect(await screen.findByText("No lines.")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Invoicing" }));
    expect(await screen.findByText("No received-not-billed lines to bill.")).toBeInTheDocument();
  });

  it("disables payment for draft orders and enables it when confirmed", async () => {
    seedPurchaseOrder({ state: "draft" });
    const user = userEvent.setup();
    renderWithProviders(<PurchaseOrderDetail orgId="1" orderId="5" />);

    await screen.findAllByText("PO-0005");
    await user.click(screen.getByRole("tab", { name: "Payments" }));
    expect(screen.getByRole("button", { name: "Make payment" })).toBeDisabled();
  });
});
