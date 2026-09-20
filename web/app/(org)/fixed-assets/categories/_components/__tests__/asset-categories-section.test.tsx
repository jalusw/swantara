import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AssetCategoriesSection } from "../asset-categories-section";

function useLocalCategories() {
  server.use(
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
            {
              id: 2,
              organization_id: 1,
              name: "Office Furniture",
              asset_account_id: null,
              depreciation_account_id: null,
              expense_account_id: null,
              gain_account_id: null,
              loss_account_id: null,
              method: "linear",
              method_number: 60,
              method_period: "month",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalCategories();
});

describe("AssetCategoriesSection", () => {
  it("renders seeded asset categories", async () => {
    renderWithProviders(<AssetCategoriesSection orgId="1" />);

    expect(await screen.findByText("IT Equipment")).toBeInTheDocument();
    expect(screen.getByText("Office Furniture")).toBeInTheDocument();
  });

  it("filters categories by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AssetCategoriesSection orgId="1" />);

    await screen.findByText("IT Equipment");
    await user.type(screen.getByPlaceholderText("Search categories…"), "Furniture");

    expect(await screen.findByText("Office Furniture")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("IT Equipment")).not.toBeInTheDocument());
  });
});
