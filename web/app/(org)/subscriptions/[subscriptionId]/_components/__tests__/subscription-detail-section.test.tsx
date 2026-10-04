import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SubscriptionDetail } from "../subscription-detail-section";

const subscription = {
  id: 5,
  organization_id: 1,
  name: "Acme Monthly",
  contact_id: 3,
  plan_id: 2,
  price_book_id: null,
  currency_code: "USD",
  date_start: "2026-01-05",
  next_invoice_date: "2026-02-05",
  date_end: null,
  state: "active",
  mrr: 99,
  lines: [{ id: 21, item_id: 7, qty: 2, unit_price: 50, discount_pct: 10 }],
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/subscriptions/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { subscription } }),
    ),
  );
});

describe("SubscriptionDetail", () => {
  it("renders the seeded subscription", async () => {
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    expect((await screen.findAllByText("Acme Monthly")).length).toBeGreaterThan(0);
  });

  it("switches to the lines tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    await screen.findAllByText("Acme Monthly");
    await user.click(screen.getByRole("tab", { name: "Baris" }));

    expect(await screen.findByText("Jml")).toBeInTheDocument();
    expect(screen.getByText("#7")).toBeInTheDocument();
  });
});
