import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ShipmentDetail } from "../shipment-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function shipment(patch: Record<string, unknown> = {}) {
  return {
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
    ...patch,
  };
}

function move(id: number, patch: Record<string, unknown> = {}) {
  return {
    id,
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
    ...patch,
  };
}

function seedShipment(shipment: unknown, moves: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/shipments/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipment } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-movements", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { movements: moves } }),
    ),
  );
}

beforeEach(() => {
  seedShipment(shipment(), [move(41)]);
});

describe("ShipmentDetail extra2", () => {
  it("shows the not-found state for a missing shipment", async () => {
    seedShipment(null, []);
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="99" />);

    expect(await screen.findByText("Shipment not found.")).toBeInTheDocument();
  });

  it("validates a draft shipment and refetches", async () => {
    let validateCalls = 0;
    server.use(
      http.get("*/api/v1/organizations/:organizationId/shipments/:id", () => {
        validateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: { shipment: shipment() } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    validateCalls = 0;
    await user.click(screen.getByRole("button", { name: "Validate" }));

    await waitFor(() => expect(validateCalls).toBeGreaterThanOrEqual(1));
  });

  it("marks an assigned shipment as done", async () => {
    let doneCalls = 0;
    seedShipment(shipment({ state: "assigned" }), [move(41, { state: "done" })]);
    server.use(
      http.get("*/api/v1/organizations/:organizationId/shipments/:id", () => {
        doneCalls += 1;
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { shipment: shipment({ state: "assigned" }) },
        });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    expect(screen.queryByRole("button", { name: "Validate" })).not.toBeInTheDocument();
    doneCalls = 0;
    await user.click(screen.getByRole("button", { name: "Mark as done" }));

    await waitFor(() => expect(doneCalls).toBeGreaterThanOrEqual(1));
  });

  it("shows no action buttons for a done shipment", async () => {
    seedShipment(shipment({ state: "done" }), []);
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    expect(screen.queryByRole("button", { name: "Validate" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Mark as done" })).not.toBeInTheDocument();
  });

  it("hides the origin row when the shipment has no origin", async () => {
    seedShipment(shipment({ origin: null }), []);
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    expect(screen.queryByText("PO-0005")).not.toBeInTheDocument();
  });

  it("shows the empty moves state", async () => {
    seedShipment(shipment(), []);
    const user = userEvent.setup();
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    await user.click(screen.getByRole("tab", { name: "Movement lines" }));

    expect(await screen.findByText("No movement lines for this shipment.")).toBeInTheDocument();
  });

  it("renders cancelled move badges", async () => {
    seedShipment(shipment(), [move(41, { state: "cancelled" }), move(42, { state: "draft" })]);
    const user = userEvent.setup();
    renderWithProviders(<ShipmentDetail orgId="1" shipmentId="21" />);

    await screen.findAllByText("PK-0021");
    await user.click(screen.getByRole("tab", { name: "Movement lines" }));

    expect(await screen.findByText("cancelled")).toBeInTheDocument();
    expect(screen.getAllByText("draft").length).toBeGreaterThan(1);
  });
});
