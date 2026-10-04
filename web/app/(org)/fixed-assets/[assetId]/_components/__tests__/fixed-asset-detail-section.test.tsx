import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { FixedAssetDetail } from "../fixed-asset-detail-section";

function useLocalAsset() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/fixed-assets/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          fixed_asset: {
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
          },
          lines: [],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalAsset();
});

describe("FixedAssetDetail", () => {
  it("renders asset name with purchase value", async () => {
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    expect(await screen.findByRole("heading", { name: "Laptop ThinkPad" })).toBeInTheDocument();
    expect(screen.getAllByText("Rp 1.500,00").length).toBeGreaterThan(0);
  });

  it("switches to depreciation schedule tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetDetail orgId="1" assetId="1" />);

    await screen.findByRole("heading", { name: "Laptop ThinkPad" });
    await user.click(screen.getByRole("tab", { name: "Jadwal penyusutan" }));

    expect(
      await screen.findByText("Belum ada jadwal penyusutan. Buat jadwal untuk memulai."),
    ).toBeInTheDocument();
  });
});
