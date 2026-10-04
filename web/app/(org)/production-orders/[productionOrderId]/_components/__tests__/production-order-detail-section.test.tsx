import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MoDetailSection } from "../production-order-detail-section";

let confirmCalled = false;

function useLocalMo() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/production-orders/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          production_order: {
            id: 1,
            organization_id: 1,
            name: "MO-0001 Chair Assembly",
            item_id: 5,
            recipe_id: null,
            qty_to_produce: 100,
            qty_produced: 0,
            unit_id: null,
            src_location_id: null,
            dst_location_id: null,
            state: "draft",
            date_planned_start: null,
            date_planned_finish: null,
            date_start: null,
            date_finished: null,
            origin: "manual",
            priority: 1,
          },
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/production-orders/:id/components", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { components: [] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/production-orders/:id/shop-tasks", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { shop_tasks: [] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/production-orders/:id/confirm", () => {
      confirmCalled = true;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          production_order: {
            id: 1,
            organization_id: 1,
            name: "MO-0001 Chair Assembly",
            item_id: 5,
            qty_to_produce: 100,
            qty_produced: 0,
            state: "confirmed",
          },
        },
      });
    }),
  );
}

beforeEach(() => {
  confirmCalled = false;
  useLocalMo();
});

describe("MoDetailSection", () => {
  it("renders the manufacturing order name", async () => {
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    expect(await screen.findByText("MO-0001 Chair Assembly")).toBeInTheDocument();
  });

  it("confirms a draft order", async () => {
    const user = userEvent.setup();
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    await screen.findByText("MO-0001 Chair Assembly");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(confirmCalled).toBe(true));
  });
});
