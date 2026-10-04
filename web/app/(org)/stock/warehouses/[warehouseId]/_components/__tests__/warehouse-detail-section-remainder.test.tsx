import { screen, waitFor, within } from "@testing-library/react";
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
  {
    id: 10,
    warehouse_id: 1,
    name: "WH/Stock/Shelf A",
    code: null,
    parent_id: 9,
    usage: "internal",
    barcode: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 11,
    warehouse_id: 2,
    name: "Other WH/Stock",
    code: "OTH",
    parent_id: null,
    usage: "internal",
    barcode: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function useLocationsHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/stock-locations", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { locations } }),
    ),
  );
}

beforeEach(() => {
  useLocationsHandlers();
});

describe("WarehouseDetail remainder", () => {
  it("renders nested locations and filters other warehouses", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    expect(await screen.findByText("WH/Stock")).toBeInTheDocument();
    expect(screen.queryByText("Other WH/Stock")).not.toBeInTheDocument();

    await user.click(screen.getByRole("treeitem", { name: /WH\/Stock/ }));
    expect(await screen.findByText("WH/Stock/Shelf A")).toBeInTheDocument();
  });

  it("renders a dash-free badge for locations without a code", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    await screen.findByText("WH/Stock");
    await user.click(screen.getByRole("treeitem", { name: /WH\/Stock/ }));
    await screen.findByText("WH/Stock/Shelf A");
    expect(screen.getByText("STK")).toBeInTheDocument();
  });

  it("shows the empty state when the warehouse has no locations", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/stock-locations", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { locations: [] } }),
      ),
    );
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    expect(
      await screen.findByText("Belum ada lokasi. Tambah lokasi untuk mengatur gudang ini."),
    ).toBeInTheDocument();
  });

  it("opens the edit dialog for a location", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    await screen.findByText("WH/Stock");
    await user.click(screen.getAllByRole("button", { name: "Ubah" })[0]!);

    expect(await screen.findByText("Ubah lokasi")).toBeInTheDocument();
  });

  it("adds a sub-location from the tree row action", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    await screen.findByText("WH/Stock");
    await user.click(screen.getAllByRole("button", { name: "Tambah sub-lokasi" })[0]!);

    expect(await screen.findByText("Lokasi baru")).toBeInTheDocument();
  });

  it("deletes a location through the tree row action", async () => {
    let deletedId: number | null = null;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/stock-locations/:id", ({ params }) => {
        deletedId = Number(params.id);
        return HttpResponse.json({ success: true, message: "Deleted." });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    await screen.findByText("WH/Stock");
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);

    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deletedId).toBe(9));
  });

  it("shows an error toast when deleting fails", async () => {
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/stock-locations/:id", () =>
        HttpResponse.json({ success: false, message: "Cannot delete." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<WarehouseDetail orgId="1" warehouseId="1" />);

    await screen.findByText("WH/Stock");
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);

    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(screen.getByText("WH/Stock")).toBeInTheDocument());
  });
});
