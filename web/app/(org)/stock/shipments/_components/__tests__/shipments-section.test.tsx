import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ShipmentsSection } from "../shipments-section";

const STAMP = "2026-01-01T00:00:00Z";

const shipments = [
  {
    id: 21,
    organization_id: 1,
    name: "PK-0001",
    type: "incoming",
    contact_id: null,
    src_location_id: null,
    dst_location_id: null,
    state: "draft",
    scheduled_date: "2026-02-01",
    date_done: null,
    origin: "PO-0005",
    carrier_id: null,
    tracking_ref: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 22,
    organization_id: 1,
    name: "PK-0002",
    type: "outgoing",
    contact_id: null,
    src_location_id: null,
    dst_location_id: null,
    state: "assigned",
    scheduled_date: "2026-02-02",
    date_done: null,
    origin: "SO-0007",
    carrier_id: null,
    tracking_ref: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/shipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { shipments } }),
    ),
  );
});

describe("ShipmentsSection", () => {
  it("renders seeded shipments", async () => {
    renderWithProviders(<ShipmentsSection />);

    expect(await screen.findByText("PK-0001")).toBeInTheDocument();
    expect(screen.getByText("PK-0002")).toBeInTheDocument();
  });

  it("filters rows through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ShipmentsSection />);

    await screen.findByText("PK-0001");
    await user.type(screen.getByPlaceholderText("Cari pengiriman…"), "PK-0002");

    expect((await screen.findAllByText("PK-0002")).length).toBeGreaterThan(0);
    expect(screen.queryByText("PK-0001")).not.toBeInTheDocument();
  });
});
