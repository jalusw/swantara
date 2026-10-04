import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { GiftCardDetail } from "../gift-card-detail-section";

const BASE_CARD = {
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

function useHandlers(options?: {
  card?: unknown;
  transactions?: unknown[];
  failActions?: boolean;
}) {
  const actions: string[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/gift-cards/:giftCardId", () => {
      if (options?.card === null) {
        return HttpResponse.json({ success: false, message: "Missing." }, { status: 404 });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { gift_card: options?.card ?? BASE_CARD },
      });
    }),
    http.get("*/api/v1/organizations/:organizationId/gift-cards/:giftCardId/transactions", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { gift_card_transactions: options?.transactions ?? [] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/gift-cards/:giftCardId/redeem", async () => {
      actions.push("redeem");
      if (options?.failActions) {
        return HttpResponse.json({ success: false, message: "Nope." }, { status: 422 });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
    http.post("*/api/v1/organizations/:organizationId/gift-cards/:giftCardId/refund", async () => {
      actions.push("refund");
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
  );
  return { actions };
}

beforeEach(() => {});

describe("GiftCardDetail branches", () => {
  it("renders fallbacks for missing relations", async () => {
    useHandlers();
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);
    expect(await screen.findByRole("heading", { name: "GC-1001" })).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders the not-found state when the card is missing", async () => {
    useHandlers({ card: null });
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="9" />);
    expect(await screen.findByText("Kartu hadiah tidak ditemukan.")).toBeInTheDocument();
  });

  it("renders transactions with order fallbacks", async () => {
    const user = userEvent.setup();
    useHandlers({
      transactions: [
        { id: 1, type: "redeem", amount: 25000, order_type: "pos", order_id: 12 },
        { id: 2, type: "refund", amount: 5000, order_type: null, order_id: null },
      ],
    });
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);
    await screen.findByRole("heading", { name: "GC-1001" });
    await user.click(screen.getByRole("tab", { name: "Transaksi" }));
    expect(await screen.findByText("#12")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("redeems an active card", async () => {
    const user = userEvent.setup();
    const { actions } = useHandlers();
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);
    await screen.findByRole("heading", { name: "GC-1001" });
    await user.click(screen.getByRole("button", { name: "Redeem" }));
    await waitFor(() => expect(actions).toEqual(["redeem"]));
  });

  it("refunds a used card without redeem action", async () => {
    const user = userEvent.setup();
    const { actions } = useHandlers({ card: { ...BASE_CARD, state: "used" } });
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);
    await screen.findByRole("heading", { name: "GC-1001" });
    expect(screen.queryByRole("button", { name: "Redeem" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Pengembalian dana" }));
    await waitFor(() => expect(actions).toEqual(["refund"]));
  });

  it("hides actions for expired cards", async () => {
    useHandlers({ card: { ...BASE_CARD, state: "expired" } });
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);
    await screen.findByRole("heading", { name: "GC-1001" });
    expect(screen.queryByRole("button", { name: "Redeem" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Pengembalian dana" })).not.toBeInTheDocument();
  });

  it("renders populated relations", async () => {
    useHandlers({
      card: {
        ...BASE_CARD,
        contact_id: 4,
        expiry_date: "2026-12-31",
        issued_from_order_id: 77,
      },
    });
    renderWithProviders(<GiftCardDetail orgId="1" giftCardId="1" />);
    await screen.findByRole("heading", { name: "GC-1001" });
    expect(screen.getByText("#4")).toBeInTheDocument();
    expect(screen.getByText("#77")).toBeInTheDocument();
  });
});
