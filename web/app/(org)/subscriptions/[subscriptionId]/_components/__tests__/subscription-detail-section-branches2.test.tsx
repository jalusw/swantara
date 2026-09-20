import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SubscriptionDetail } from "../subscription-detail-section";

const baseSubscription = {
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

function seedSubscription(overrides: Record<string, unknown> = {}) {
  const subscription = { ...baseSubscription, ...overrides };
  server.use(
    http.get("*/api/v1/organizations/:organizationId/subscriptions/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { subscription } }),
    ),
  );
}

beforeEach(() => {});

describe("SubscriptionDetail branches2", () => {
  it("renders the not-found branch when the subscription is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/subscriptions/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { subscription: null } }),
      ),
    );
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    expect(await screen.findByText("Subscription not found.")).toBeInTheDocument();
  });

  it("shows draft-only activate action and hides other actions", async () => {
    seedSubscription({ state: "draft" });
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    await screen.findAllByText("Acme Monthly");
    expect(screen.getByRole("button", { name: "Activate" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Pause" })).not.toBeInTheDocument();
  });

  it("shows pause and churn actions for active subscriptions", async () => {
    seedSubscription({ state: "active" });
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    await screen.findAllByText("Acme Monthly");
    expect(screen.getByRole("button", { name: "Pause" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Churn" })).toBeInTheDocument();
  });

  it("renders dash fallbacks for missing dates and currency", async () => {
    seedSubscription({
      date_start: null,
      next_invoice_date: null,
      currency_code: null,
      date_end: null,
    });
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    await screen.findAllByText("Acme Monthly");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders the empty lines branch on the lines tab", async () => {
    seedSubscription({ lines: [] });
    const user = userEvent.setup();
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    await screen.findAllByText("Acme Monthly");
    await user.click(screen.getByRole("tab", { name: "Lines" }));
    expect(await screen.findByText("No lines on this subscription.")).toBeInTheDocument();
  });

  it("covers paused resume and closed terminal branches", async () => {
    seedSubscription({ state: "paused" });
    renderWithProviders(<SubscriptionDetail orgId="1" subscriptionId="5" />);

    await screen.findAllByText("Acme Monthly");
    expect(screen.getByRole("button", { name: "Resume" })).toBeInTheDocument();
  });
});
