import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TransfersSection } from "../transfers-section";

const STAMP = "2026-01-01T00:00:00Z";

const orders = [
  {
    id: 31,
    organization_id: 1,
    name: "TO-0001",
    src_warehouse_id: 1,
    dst_warehouse_id: 2,
    state: "draft",
    out_shipment_id: null,
    in_shipment_id: null,
    is_interorganization: false,
    scheduled_date: "2026-02-01",
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 32,
    organization_id: 1,
    name: "TO-0002",
    src_warehouse_id: 1,
    dst_warehouse_id: 2,
    state: "in_transit",
    out_shipment_id: 21,
    in_shipment_id: null,
    is_interorganization: false,
    scheduled_date: "2026-02-02",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/warehouse-transfers", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouseTransfers: orders } }),
    ),
  );
});

describe("TransfersSection", () => {
  it("renders seeded transfer orders", async () => {
    renderWithProviders(<TransfersSection />);

    expect(await screen.findByText("TO-0001")).toBeInTheDocument();
    expect(screen.getByText("TO-0002")).toBeInTheDocument();
  });

  it("filters rows through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TransfersSection />);

    await screen.findByText("TO-0001");
    await user.type(screen.getByPlaceholderText("Cari transfer…"), "TO-0002");

    expect((await screen.findAllByText("TO-0002")).length).toBeGreaterThan(0);
    expect(screen.queryByText("TO-0001")).not.toBeInTheDocument();
  });
});
