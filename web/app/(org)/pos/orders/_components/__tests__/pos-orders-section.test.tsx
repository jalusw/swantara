import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosOrdersSection } from "../pos-orders-section";

const orders = [
  {
    id: 1,
    session_id: 1,
    contact_id: null,
    name: "Walk-in Sale",
    amount_total: 250,
    amount_tax: 25,
    state: "done",
    invoice_id: null,
    order_time: "2026-01-01T10:00:00Z",
  },
  {
    id: 2,
    session_id: 2,
    contact_id: 2,
    name: "Corporate Order",
    amount_total: 1200,
    amount_tax: 120,
    state: "refunded",
    invoice_id: 7,
    order_time: "2026-01-02T11:00:00Z",
  },
];

function useOrderHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/orders", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { orders } }),
    ),
  );
}

beforeEach(() => {
  useOrderHandlers();
});

describe("PosOrdersSection", () => {
  it("renders orders with session links and state badges", async () => {
    renderWithProviders(<PosOrdersSection />);

    expect(await screen.findByText("Walk-in Sale")).toBeInTheDocument();
    expect(screen.getByText("Corporate Order")).toBeInTheDocument();
    expect(screen.getByText("Session-1")).toBeInTheDocument();
    expect(screen.getByText("Session-2")).toBeInTheDocument();
    expect(screen.getByText("#7")).toBeInTheDocument();
  });

  it("filters orders through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosOrdersSection />);

    await screen.findByText("Walk-in Sale");
    await user.type(screen.getByPlaceholderText("Search orders..."), "Corporate");

    expect(await screen.findByText("Corporate Order")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Walk-in Sale")).not.toBeInTheDocument());
  });
});
