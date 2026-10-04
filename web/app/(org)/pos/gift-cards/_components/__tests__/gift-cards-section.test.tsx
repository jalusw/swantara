import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { GiftCardsSection } from "../gift-cards-section";

const giftCards = [
  {
    id: 1,
    organization_id: 1,
    code: "GC-1001",
    contact_id: null,
    initial_amount: 100000,
    balance: 75000,
    currency_code: "USD",
    expiry_date: null,
    state: "active",
    issued_from_order_id: null,
  },
  {
    id: 2,
    organization_id: 1,
    code: "GC-2002",
    contact_id: null,
    initial_amount: 50000,
    balance: 0,
    currency_code: "USD",
    expiry_date: null,
    state: "used",
    issued_from_order_id: null,
  },
];

function useGiftCardHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/gift-cards", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { gift_cards: giftCards } }),
    ),
  );
}

beforeEach(() => {
  useGiftCardHandlers();
});

describe("GiftCardsSection", () => {
  it("renders cards with state badges", async () => {
    renderWithProviders(<GiftCardsSection orgId="1" />);

    expect(await screen.findByText("GC-1001")).toBeInTheDocument();
    expect(screen.getByText("GC-2002")).toBeInTheDocument();
    expect(screen.getByText("Aktif")).toBeInTheDocument();
    expect(screen.getByText("Terpakai")).toBeInTheDocument();
  });

  it("filters cards through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<GiftCardsSection orgId="1" />);

    await screen.findByText("GC-1001");
    await user.type(screen.getByPlaceholderText(/Cari kartu hadiah/), "2002");

    expect(await screen.findByText("GC-2002")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("GC-1001")).not.toBeInTheDocument());
  });
});
