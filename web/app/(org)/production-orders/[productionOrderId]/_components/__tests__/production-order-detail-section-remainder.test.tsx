import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MoDetailSection } from "../production-order-detail-section";

function moFixture(overrides = {}) {
  return {
    id: 1,
    organization_id: 1,
    name: "MO-0001 Chair Assembly",
    item_id: 5,
    recipe_id: null,
    qty_to_produce: 100,
    qty_produced: 25,
    unit_id: null,
    src_location_id: null,
    dst_location_id: null,
    state: "draft",
    date_planned_start: "2026-02-01T08:00:00Z",
    date_planned_finish: "2026-02-05T17:00:00Z",
    date_start: null,
    date_finished: null,
    origin: "manual",
    priority: 1,
    ...overrides,
  };
}

function useMoHandlers(
  productionOrder: Record<string, unknown>,
  components: Record<string, unknown>[] = [],
  shopTasks: Record<string, unknown>[] = [],
) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/production-orders/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { production_order: productionOrder },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/production-orders/:id/components", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { components } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/production-orders/:id/shop-tasks", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shop_tasks: shopTasks } }),
    ),
  );
}

beforeEach(() => {});

describe("MoDetailSection remainder", () => {
  it("shows the not-found state when the order is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/production-orders/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="999" />);

    expect(await screen.findByText("Manufacturing order not found")).toBeInTheDocument();
  });

  it("renders components with consumption progress", async () => {
    useMoHandlers(moFixture(), [
      {
        id: 3,
        production_order_id: 1,
        item_id: 9,
        qty_planned: 200,
        qty_consumed: 50,
      },
    ]);
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    expect(await screen.findByText("MO-0001 Chair Assembly")).toBeInTheDocument();
    expect(screen.getByText("#9")).toBeInTheDocument();
    expect(screen.getAllByText("25%").length).toBeGreaterThanOrEqual(2);
  });

  it("renders work orders with fallback names", async () => {
    useMoHandlers(
      moFixture(),
      [],
      [
        {
          id: 7,
          production_order_id: 1,
          name: null,
          work_center_id: null,
          planned_minutes: null,
          actual_minutes: 30,
          state: "in_progress",
        },
      ],
    );
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    expect(await screen.findByText("WO-7")).toBeInTheDocument();
    expect(screen.getByText("in_progress")).toBeInTheDocument();
  });

  it("plans a confirmed order", async () => {
    let planned = false;
    useMoHandlers(moFixture({ state: "confirmed" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/production-orders/:id/plan", () => {
        planned = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    await screen.findByText("MO-0001 Chair Assembly");
    await user.click(screen.getByRole("button", { name: "Plan" }));

    await waitFor(() => expect(planned).toBe(true));
  });

  it("starts a planned order", async () => {
    let started = false;
    useMoHandlers(moFixture({ state: "planned" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/production-orders/:id/start", () => {
        started = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    await screen.findByText("MO-0001 Chair Assembly");
    await user.click(screen.getByRole("button", { name: "Start" }));

    await waitFor(() => expect(started).toBe(true));
  });

  it("opens consume and produce dialogs for in-progress orders", async () => {
    useMoHandlers(moFixture({ state: "in_progress" }));
    const user = userEvent.setup();
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    await screen.findByText("MO-0001 Chair Assembly");
    const actionButtons = screen
      .getAllByRole("button")
      .filter((button) => /consume/i.test(button.textContent ?? ""));
    expect(actionButtons.length).toBeGreaterThanOrEqual(1);
    await user.click(actionButtons[0]!);
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("cancels an open order", async () => {
    let cancelled = false;
    useMoHandlers(moFixture({ state: "confirmed" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/production-orders/:id/cancel", () => {
        cancelled = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<MoDetailSection orgId="1" productionOrderId="1" />);

    await screen.findByText("MO-0001 Chair Assembly");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(cancelled).toBe(true));
  });
});
