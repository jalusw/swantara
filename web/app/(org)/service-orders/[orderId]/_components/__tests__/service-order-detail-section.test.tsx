import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceOrderDetail } from "../service-order-detail-section";

function useLocalOrder() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/service-orders/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          service_order: {
            id: 1,
            organization_id: 1,
            name: "AC Repair Visit",
            contact_id: null,
            equipment_id: null,
            contract_id: null,
            type: "repair",
            priority: 1,
            state: "new",
            scheduled_date: null,
            technician_id: null,
            invoice_id: null,
            dimension_id: null,
            reported_issue: "AC not cooling",
            resolution: "",
          },
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/service-orders/:id/lines", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          service_order_lines: [
            {
              id: 1,
              service_order_id: 1,
              type: "labor",
              item_id: null,
              description: "Brake pad replacement",
              qty: 2,
              unit_id: null,
              unit_cost: 10,
              unit_price: 25,
              stock_movement_id: null,
              billable: true,
              covered_by_warranty: false,
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalOrder();
});

describe("ServiceOrderDetail", () => {
  it("renders the service order name", async () => {
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    expect(await screen.findByRole("heading", { name: "AC Repair Visit" })).toBeInTheDocument();
  });

  it("switches to the lines tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    await user.click(screen.getByRole("tab", { name: "Baris" }));

    expect(await screen.findByText("Brake pad replacement")).toBeInTheDocument();
  });
});
