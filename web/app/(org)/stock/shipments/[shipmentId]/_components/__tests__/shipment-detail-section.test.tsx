import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ShipmentDetail } from "../shipment-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const shipment = {
  id: 21,
  organization_id: 1,
  name: "PK-0021",
  type: "incoming",
  contact_id: null,
  src_location_id: 9,
  dst_location_id: 10,
  state: "draft",
  scheduled_date: "2026-02-01",
  date_done: null,
  origin: "PO-0005",
  carrier_id: null,
  tracking_ref: null,
  created_at: STAMP,
  updated_at: STAMP,
};

const moves = [
  {
    id: 41,
    organization_id: 1,
    shipment_id: 21,
    item_id: 5,
    qty: 10,
    unit_id: null,
    src_location_id: 9,
    dst_location_id: 10,
    batch_id: null,
    state: "assigned",
    unit_cost: null,
    origin_type: null,
    origin_id: null,
    scheduled_date: null,
    date_done: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/shipments/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipment } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-movements", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { movements: moves } }),
    ),
  );
});

describe("ShipmentDetail", () => {
  it("renders the shipment header with origin", async () => {
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    expect((await screen.findAllByText("PK-0021")).length).toBeGreaterThan(0);
    expect(screen.getByText("PO-0005")).toBeInTheDocument();
  });

  it("switches to the moves tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    await user.click(screen.getByRole("tab", { name: "Movement lines" }));

    expect(await screen.findByText("assigned")).toBeInTheDocument();
  });
});
