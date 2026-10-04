import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { DashboardTabs } from "../dashboard-tabs";

function useDashboardTabHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/kpis/payroll", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { kpi: { gross_cost: 45000, net_cost: 38000 } },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/kpis/projects", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          kpi: {
            project_count: 12,
            total_margin: 85000,
            total_cost: 120000,
            total_billed: 205000,
            utilization: 0.75,
          },
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/kpis/subscription", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { kpi: { mrr: 12500, arr: 150000, churned: 3, churn_rate: 0.02, ltv: 45000 } },
      }),
    ),
  );
}

beforeEach(() => {
  useDashboardTabHandlers();
});

describe("DashboardTabs", () => {
  it("renders tab triggers for every dashboard area", async () => {
    renderWithProviders(<DashboardTabs orgId="1" />);

    expect(await screen.findByRole("tab", { name: "Ikhtisar" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Operasi" })).toBeInTheDocument();
  });

  it("switches to the operations tab with payroll, project and subscription sections", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardTabs orgId="1" />);

    await user.click(await screen.findByRole("tab", { name: "Operasi" }));

    expect(await screen.findByText("Penggajian")).toBeInTheDocument();
    expect(screen.getAllByText("Biaya bruto").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Proyek").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Langganan").length).toBeGreaterThan(0);
  });
});
