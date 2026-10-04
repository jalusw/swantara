import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosSessionDetail } from "../pos-session-detail-section";

const session = {
  id: 1,
  config_id: 5,
  cashier_id: 7,
  opened_at: "2026-01-01T08:00:00Z",
  closed_at: null,
  opening_balance: 100000,
  closing_balance: null,
  state: "opened",
  orders: [
    {
      id: 11,
      session_id: 1,
      contact_id: null,
      name: "POS-001",
      amount_total: 50000,
      amount_tax: 5000,
      state: "done",
      invoice_id: null,
      order_time: "2026-01-02T10:00:00Z",
      lines: [],
      payments: [],
    },
  ],
};

function usePosSessionDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/sessions/:sessionId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { session } }),
    ),
  );
}

beforeEach(() => {
  usePosSessionDetailHandlers();
});

describe("PosSessionDetail", () => {
  it("renders session with balances", async () => {
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    expect(await screen.findByRole("heading", { name: "Sesi 1" })).toBeInTheDocument();
    expect(screen.getByText("Total penjualan")).toBeInTheDocument();
  });

  it("shows orders on the orders tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    await user.click(screen.getByRole("tab", { name: "Pesanan" }));

    expect(await screen.findByText("POS-001")).toBeInTheDocument();
  });
});
