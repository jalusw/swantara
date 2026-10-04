import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SubscriptionPlansSection } from "../subscription-plans-section";

const plans = [
  { id: 1, name: "Monthly Basic", recurring_interval: "monthly", recurring_count: 1 },
  { id: 2, name: "Annual Pro", recurring_interval: "yearly", recurring_count: 1 },
];

function useSubscriptionPlanHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/subscription-plans", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { plans } }),
    ),
  );
}

beforeEach(() => {
  useSubscriptionPlanHandlers();
});

describe("SubscriptionPlansSection", () => {
  it("renders plans with intervals", async () => {
    renderWithProviders(<SubscriptionPlansSection orgId="1" />);

    expect(await screen.findByText("Monthly Basic")).toBeInTheDocument();
    expect(screen.getByText("Annual Pro")).toBeInTheDocument();
  });

  it("filters plans through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SubscriptionPlansSection orgId="1" />);

    await screen.findByText("Monthly Basic");
    await user.type(screen.getByPlaceholderText(/Cari paket/), "Annual");

    expect(await screen.findByText("Annual Pro")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Monthly Basic")).not.toBeInTheDocument());
  });
});
