import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { FixedAssetDetail } from "../fixed-asset-detail-section";

function asset(patch: Record<string, unknown> = {}) {
  return {
    id: 1,
    organization_id: 1,
    name: "Laptop ThinkPad",
    category_id: 2,
    purchase_value: 1500,
    salvage_value: 100,
    acquisition_date: "2025-06-01",
    in_service_date: "2025-06-05",
    original_entry_id: null,
    invoice_line_id: null,
    state: "draft",
    disposal_date: null,
    ...patch,
  };
}

function seedAsset(assetData: unknown, lines: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/fixed-assets/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { fixed_asset: assetData, lines },
      }),
    ),
  );
}

beforeEach(() => {
  seedAsset(asset());
});

describe("FixedAssetDetail extra2", () => {
  it("shows the not-found state for a missing asset", async () => {
    seedAsset(null);
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    expect(await screen.findByText("Aset tidak ditemukan.")).toBeInTheDocument();
  });

  it("generates the depreciation schedule for a running asset", async () => {
    let scheduleCalls = 0;
    seedAsset(asset({ state: "running" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/fixed-assets/:id/schedule", () => {
        scheduleCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: { lines: [] } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    await user.click(screen.getByRole("button", { name: "Buat jadwal" }));

    await waitFor(() => expect(scheduleCalls).toBe(1));
  });

  it("posts depreciation when unposted lines exist", async () => {
    let postCalls = 0;
    seedAsset(asset({ state: "running" }), [
      {
        id: 1,
        sequence: 1,
        depreciation_date: "2025-07-01",
        amount: 100,
        accumulated: 100,
        remaining_value: 1400,
        posted: false,
      },
    ]);
    server.use(
      http.post("*/api/v1/organizations/:organizationId/fixed-assets/:id/post-depreciation", () => {
        postCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    await user.click(screen.getByRole("button", { name: "Posting penyusutan" }));

    await waitFor(() => expect(postCalls).toBe(1));
  });

  it("renders the schedule table with posted and pending badges", async () => {
    seedAsset(asset({ state: "running" }), [
      {
        id: 1,
        sequence: 1,
        depreciation_date: "2025-07-01",
        amount: 100,
        accumulated: 100,
        remaining_value: 1400,
        posted: true,
      },
      {
        id: 2,
        sequence: 2,
        depreciation_date: "2025-08-01",
        amount: 100,
        accumulated: 200,
        remaining_value: 1300,
        posted: false,
      },
    ]);
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    await user.click(screen.getByRole("tab", { name: "Jadwal penyusutan" }));

    expect(await screen.findByText("Periode berikutnya untuk diposting")).toBeInTheDocument();
    expect(screen.getAllByText("Diposting").length).toBeGreaterThan(1);
    expect(screen.getByText("Menunggu")).toBeInTheDocument();
  });

  it("disposes the asset through the dispose dialog", async () => {
    let disposeCalls = 0;
    seedAsset(asset({ state: "running" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/fixed-assets/:id/dispose", () => {
        disposeCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    await user.click(screen.getByRole("button", { name: "Lepas / jual" }));
    expect(await screen.findByText("Lepas / jual aset")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(disposeCalls).toBe(1));
  });

  it("falls back to dash for missing acquisition dates", async () => {
    seedAsset(asset({ state: "sold", acquisition_date: null, in_service_date: null }));
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("hides state actions for a disposed asset", async () => {
    seedAsset(asset({ state: "disposed" }));
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    expect(screen.queryByRole("button", { name: "Buat jadwal" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Posting penyusutan" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Lepas / jual" })).not.toBeInTheDocument();
  });
});
