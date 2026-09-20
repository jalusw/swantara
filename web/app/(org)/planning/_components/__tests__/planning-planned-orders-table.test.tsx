import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PlannedSupplysTable } from "../planning-planned-orders-table";

const STAMP = "2026-01-05T00:00:00Z";

const runDetail = {
  id: 1,
  organization_id: 1,
  run_date: STAMP,
  horizon_days: 30,
  state: "done",
  created_at: STAMP,
  updated_at: STAMP,
  planned_orders: [
    {
      id: 11,
      planning_run_id: 1,
      item_id: 7,
      warehouse_id: null,
      type: "purchase",
      qty: 50,
      order_date: "2026-01-06T00:00:00Z",
      due_date: "2026-01-20T00:00:00Z",
      pegged_demand_id: null,
      confirmed: true,
      generated_doc_type: null,
      generated_doc_id: null,
      created_at: STAMP,
      updated_at: STAMP,
    },
    {
      id: 12,
      planning_run_id: 1,
      item_id: 9,
      warehouse_id: null,
      type: "manufacture",
      qty: 20,
      order_date: null,
      due_date: null,
      pegged_demand_id: null,
      confirmed: false,
      generated_doc_type: null,
      generated_doc_id: null,
      created_at: STAMP,
      updated_at: STAMP,
    },
  ],
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/planning/runs/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { run: runDetail } }),
    ),
  );
});

describe("PlannedSupplysTable", () => {
  it("renders seeded planned orders", async () => {
    renderWithProviders(
      <PlannedSupplysTable
        run={{
          id: 1,
          organizationId: 1,
          runDate: new Date(STAMP),
          horizonDays: 30,
          state: "done",
          createdAt: new Date(STAMP),
          updatedAt: new Date(STAMP),
        }}
      />,
    );

    expect(await screen.findByText("#7")).toBeInTheDocument();
    expect(screen.getByText("#9")).toBeInTheDocument();
    expect(screen.getByText("Confirmed")).toBeInTheDocument();
    expect(screen.getByText("Pending")).toBeInTheDocument();
  });

  it("keeps orders visible when rows are clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PlannedSupplysTable
        run={{
          id: 1,
          organizationId: 1,
          runDate: new Date(STAMP),
          horizonDays: 30,
          state: "done",
          createdAt: new Date(STAMP),
          updatedAt: new Date(STAMP),
        }}
      />,
    );

    await user.click(await screen.findByText("#7"));

    expect(screen.getByText("#9")).toBeInTheDocument();
  });
});
