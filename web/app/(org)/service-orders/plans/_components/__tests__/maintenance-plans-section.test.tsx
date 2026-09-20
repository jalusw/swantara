import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MaintenancePlansSection } from "../maintenance-plans-section";

const maintenancePlans = [
  {
    id: 1,
    equipment_id: 3,
    name: "Quarterly Inspection",
    interval_days: 90,
    next_due: "2026-04-01",
    active: true,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/maintenance-plans", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { maintenance_plans: maintenancePlans },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { equipments: [] } }),
    ),
  );
});

describe("MaintenancePlansSection", () => {
  it("renders seeded maintenance plans", async () => {
    renderWithProviders(<MaintenancePlansSection orgId="1" />);

    expect(await screen.findByText("Quarterly Inspection")).toBeInTheDocument();
    expect(screen.getByText("90 days")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<MaintenancePlansSection orgId="1" />);

    await screen.findByText("Quarterly Inspection");
    await user.click(screen.getByRole("button", { name: "Create plan" }));

    expect(
      await screen.findByRole("heading", { name: "Create maintenance plan" }),
    ).toBeInTheDocument();
  });
});
