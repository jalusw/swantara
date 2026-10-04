import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { WarehouseDetail } from "../warehouse-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const locations = [
  {
    id: 9,
    warehouse_id: 1,
    name: "WH/Stock",
    code: "STK",
    parent_id: null,
    usage: "internal",
    barcode: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/stock-locations", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { locations } }),
    ),
  );
});

describe("WarehouseDetail", () => {
  it("renders seeded locations as a tree", async () => {
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    expect(await screen.findByText("WH/Stock")).toBeInTheDocument();
    expect(screen.getByText("Lokasi stok")).toBeInTheDocument();
  });

  it("opens the create location dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    await screen.findByText("WH/Stock");
    await user.click(screen.getByRole("button", { name: "Tambah lokasi" }));

    expect(await screen.findByRole("heading", { name: "Lokasi baru" })).toBeInTheDocument();
  });
});
