import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SubscriptionsSection } from "../subscriptions-section";

const subscriptions = [
  {
    id: 1,
    organization_id: 1,
    name: "Acme Monthly",
    contact_id: 1,
    plan_id: 1,
    price_book_id: 1,
    currency_code: "USD",
    date_start: "2026-01-01",
    next_invoice_date: "2026-02-01",
    date_end: null,
    state: "active",
    mrr: 5000,
  },
  {
    id: 2,
    organization_id: 1,
    name: "Beta Annual",
    contact_id: 2,
    plan_id: 1,
    price_book_id: 1,
    currency_code: "USD",
    date_start: "2026-01-01",
    next_invoice_date: "2027-01-01",
    date_end: null,
    state: "paused",
    mrr: 7500,
  },
];

const metrics = { mrr: 12500, arr: 150000, churned: 3, churn_rate: 0.02, ltv: 45000 };

function useSubscriptionHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/subscriptions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { subscriptions } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/subscriptions/metrics", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { metrics } }),
    ),
  );
}

beforeEach(() => {
  useSubscriptionHandlers();
});

describe("SubscriptionsSection", () => {
  it("renders subscriptions with metric cards and state badges", async () => {
    renderWithProviders(<SubscriptionsSection orgId="1" />);

    expect(await screen.findByText("Acme Monthly")).toBeInTheDocument();
    expect(screen.getByText("Beta Annual")).toBeInTheDocument();
    expect(screen.getAllByText("MRR").length).toBeGreaterThan(0);
    expect(screen.getByText("Aktif")).toBeInTheDocument();
    expect(screen.getByText("Dijeda")).toBeInTheDocument();
  });

  it("filters subscriptions through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SubscriptionsSection orgId="1" />);

    await screen.findByText("Acme Monthly");
    await user.type(screen.getByPlaceholderText(/Cari langganan/), "Beta");

    expect(await screen.findByText("Beta Annual")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Acme Monthly")).not.toBeInTheDocument());
  });
});
