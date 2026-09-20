import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CommissionPlanDetail } from "../commission-plan-detail-section";

const commissionPlan = {
  id: 1,
  organization_id: 1,
  name: "Retail Plan",
  basis: "revenue",
  active: true,
};

function useCommissionPlanDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/commission-plans/:planId", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { commission_plan: commissionPlan },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/commission-plans/:planId/rules", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { commission_rules: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/commission-plans/:planId/assignments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { commission_assignments: [] } }),
    ),
  );
}

beforeEach(() => {
  useCommissionPlanDetailHandlers();
});

describe("CommissionPlanDetail", () => {
  it("renders plan name with status badge", async () => {
    renderWithProviders(<CommissionPlanDetail orgId="1" planId="1" />);

    expect(await screen.findByRole("heading", { name: "Retail Plan" })).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
  });

  it("shows empty rules on the rules tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CommissionPlanDetail orgId="1" planId="1" />);

    await screen.findByRole("heading", { name: "Retail Plan" });
    await user.click(screen.getByRole("tab", { name: "Rules" }));

    expect(await screen.findByText("No rules configured.")).toBeInTheDocument();
  });
});
