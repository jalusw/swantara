import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceOrdersSection } from "../service-orders-section";

function useLocalOrders() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/service-orders", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          service_orders: [
            {
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
            {
              id: 2,
              organization_id: 1,
              name: "Elevator Maintenance",
              contact_id: null,
              equipment_id: null,
              contract_id: null,
              type: "maintenance",
              priority: 2,
              state: "scheduled",
              scheduled_date: null,
              technician_id: null,
              invoice_id: null,
              dimension_id: null,
              reported_issue: "",
              resolution: "",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalOrders();
});

describe("ServiceOrdersSection", () => {
  it("renders seeded service orders", async () => {
    renderWithProviders(<ServiceOrdersSection orgId="1" />);

    expect(await screen.findByText("AC Repair Visit")).toBeInTheDocument();
    expect(screen.getByText("Elevator Maintenance")).toBeInTheDocument();
  });

  it("filters orders by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrdersSection orgId="1" />);

    await screen.findByText("AC Repair Visit");
    await user.type(screen.getByPlaceholderText("Search service orders…"), "Elevator");

    expect(await screen.findByText("Elevator Maintenance")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("AC Repair Visit")).not.toBeInTheDocument());
  });
});
