import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CommissionPlansSection } from "../commission-plans-section";

const commissionPlans = [
  { id: 1, organization_id: 1, name: "Atlas Retail", basis: "revenue", active: true },
  { id: 2, organization_id: 1, name: "Bravo Services", basis: "margin", active: false },
];

function useCommissionHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/commission-plans", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { commission_plans: commissionPlans },
      }),
    ),
  );
}

beforeEach(() => {
  useCommissionHandlers();
});

describe("CommissionPlansSection", () => {
  it("renders plans with total count", async () => {
    renderWithProviders(<CommissionPlansSection orgId="1" />);

    expect(await screen.findByText("Atlas Retail")).toBeInTheDocument();
    expect(screen.getByText("Bravo Services")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("filters plans through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CommissionPlansSection orgId="1" />);

    await screen.findByText("Atlas Retail");
    await user.type(screen.getByPlaceholderText(/Search plans/), "Bravo");

    expect(await screen.findByText("Bravo Services")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Atlas Retail")).not.toBeInTheDocument());
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CommissionPlansSection orgId="1" />);

    await screen.findByText("Atlas Retail");
    await user.click(screen.getByRole("button", { name: "Create plan" }));

    expect(await screen.findByText("Create commission plan")).toBeInTheDocument();
  });
});
