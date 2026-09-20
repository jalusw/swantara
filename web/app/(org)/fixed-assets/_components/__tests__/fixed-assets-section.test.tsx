import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { FixedAssetsSection } from "../fixed-assets-section";

function useLocalAssets() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/fixed-assets", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          fixed_assets: [
            {
              id: 1,
              organization_id: 1,
              name: "Laptop ThinkPad",
              category_id: 1,
              purchase_value: 1500,
              salvage_value: 100,
              acquisition_date: "2025-06-01",
              in_service_date: "2025-06-05",
              original_entry_id: null,
              invoice_line_id: null,
              state: "running",
              disposal_date: null,
            },
            {
              id: 2,
              organization_id: 1,
              name: "Office Printer",
              category_id: 1,
              purchase_value: 800,
              salvage_value: 50,
              acquisition_date: "2025-07-01",
              in_service_date: "2025-07-02",
              original_entry_id: null,
              invoice_line_id: null,
              state: "draft",
              disposal_date: null,
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/asset-categories", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          asset_categories: [
            {
              id: 1,
              organization_id: 1,
              name: "IT Equipment",
              asset_account_id: null,
              depreciation_account_id: null,
              expense_account_id: null,
              gain_account_id: null,
              loss_account_id: null,
              method: "linear",
              method_number: 36,
              method_period: "month",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalAssets();
});

describe("FixedAssetsSection", () => {
  it("renders seeded assets", async () => {
    renderWithProviders(<FixedAssetsSection orgId="1" />);

    expect(await screen.findByText("Laptop ThinkPad")).toBeInTheDocument();
    expect(screen.getByText("Office Printer")).toBeInTheDocument();
  });

  it("filters assets by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetsSection orgId="1" />);

    await screen.findByText("Laptop ThinkPad");
    await user.type(screen.getByPlaceholderText("Search assets…"), "Printer");

    expect(await screen.findByText("Office Printer")).toBeInTheDocument();
  });

  it("opens the register dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FixedAssetsSection orgId="1" />);

    await screen.findByText("Laptop ThinkPad");
    await user.click(screen.getByRole("button", { name: "Register asset" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
