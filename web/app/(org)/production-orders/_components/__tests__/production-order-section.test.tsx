import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MoSection } from "../production-order-section";

function useLocalMos() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/production-orders", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          production_orders: [
            {
              id: 1,
              organization_id: 1,
              name: "MO-0001 Chair Assembly",
              item_id: 5,
              recipe_id: null,
              qty_to_produce: 100,
              qty_produced: 20,
              unit_id: null,
              src_location_id: null,
              dst_location_id: null,
              state: "in_progress",
              date_planned_start: null,
              date_planned_finish: null,
              date_start: null,
              date_finished: null,
              origin: "manual",
              priority: 1,
            },
            {
              id: 2,
              organization_id: 1,
              name: "MO-0002 Table Assembly",
              item_id: 6,
              recipe_id: null,
              qty_to_produce: 50,
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
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          products: [
            { id: 5, name: "Wooden Chair" },
            { id: 6, name: "Wooden Table" },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalMos();
});

describe("MoSection", () => {
  it("renders seeded manufacturing orders", async () => {
    renderWithProviders(<MoSection orgId="1" />);

    expect(await screen.findByText("MO-0001 Chair Assembly")).toBeInTheDocument();
    expect(screen.getByText("MO-0002 Table Assembly")).toBeInTheDocument();
  });

  it("filters orders by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<MoSection orgId="1" />);

    await screen.findByText("MO-0001 Chair Assembly");
    await user.type(screen.getByPlaceholderText("Cari perintah produksi…"), "Table");

    expect(await screen.findByText("MO-0002 Table Assembly")).toBeInTheDocument();
    await waitFor(() =>
      expect(screen.queryByText("MO-0001 Chair Assembly")).not.toBeInTheDocument(),
    );
  });
});
