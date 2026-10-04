import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MrpSection } from "../planning-section";

const STAMP = "2026-01-05T00:00:00Z";

const runs = [
  {
    id: 1,
    organization_id: 1,
    run_date: STAMP,
    horizon_days: 30,
    state: "done",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const runDetail = {
  ...runs[0],
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
    http.get("*/api/v1/organizations/:organizationId/planning/runs", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { runs } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/planning/runs/:id", ({ params }) =>
      params.id === "1"
        ? HttpResponse.json({ success: true, message: "OK.", data: { run: runDetail } })
        : HttpResponse.json({ success: false }, { status: 404 }),
    ),
  );
});

describe("MrpSection", () => {
  it("renders seeded Planning runs", async () => {
    renderWithProviders(<MrpSection orgId="1" />);

    expect(await screen.findByText("Proses 1")).toBeInTheDocument();
  });

  it("shows planned orders when a run is selected", async () => {
    const user = userEvent.setup();
    renderWithProviders(<MrpSection orgId="1" />);

    await user.click(await screen.findByText("Proses 1"));

    expect(await screen.findByText("Pesanan terencana")).toBeInTheDocument();
    expect(screen.getByText("#7")).toBeInTheDocument();
  });
});
