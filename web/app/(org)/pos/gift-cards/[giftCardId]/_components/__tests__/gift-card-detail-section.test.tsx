import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { GiftCardDetail } from "../gift-card-detail-section";

const giftCard = {
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
};

function useGiftCardDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/gift-cards/:giftCardId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { gift_card: giftCard } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/gift-cards/:giftCardId/transactions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { gift_card_transactions: [] } }),
    ),
  );
}

beforeEach(() => {
  useGiftCardDetailHandlers();
});

describe("GiftCardDetail", () => {
  it("renders card code with state badge", async () => {
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);

    expect(await screen.findByRole("heading", { name: "GC-1001" })).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
  });

  it("shows empty transactions on the transactions tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);

    await screen.findByRole("heading", { name: "GC-1001" });
    await user.click(screen.getByRole("tab", { name: "Transactions" }));

    expect(await screen.findByText("No transactions.")).toBeInTheDocument();
  });
});
