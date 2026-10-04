import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosOrderDetail } from "../pos-order-detail-section";

const order = {
  id: 10,
  session_id: 1,
  contact_id: null,
  name: "POS-001",
  amount_total: 55000,
  amount_tax: 5000,
  state: "done",
  invoice_id: null,
  order_time: "2026-01-02T10:00:00Z",
  lines: [
    {
      id: 1,
      order_id: 10,
      item_id: 5,
      qty: 2,
      unit_price: 25000,
      discount_pct: 0,
      tax_ids: [],
      price_subtotal: 50000,
      price_tax: 5000,
      price_total: 55000,
    },
  ],
  payments: [{ id: 1, order_id: 10, method: "cash", amount: 55000 }],
};

function usePosOrderDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/orders/:orderId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { order } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
  );
}

beforeEach(() => {
  usePosOrderDetailHandlers();
});

describe("PosOrderDetail", () => {
  it("renders order name with state badge", async () => {
    renderWithProviders(<PosOrderDetail orgId="1" orderId="10" />);

    expect(await screen.findByRole("heading", { name: "POS-001" })).toBeInTheDocument();
    expect(screen.getByText("Info pesanan")).toBeInTheDocument();
  });

  it("shows payments on the payments tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosOrderDetail orgId="1" orderId="10" />);

    await screen.findByRole("heading", { name: "POS-001" });
    await user.click(screen.getByRole("tab", { name: "Pembayaran" }));

    expect(await screen.findByText("Tunai")).toBeInTheDocument();
  });
});
