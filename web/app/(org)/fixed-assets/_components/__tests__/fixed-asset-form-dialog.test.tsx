import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { FixedAssetFormDialog } from "../fixed-asset-form-dialog";

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
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalCategories();
});

describe("FixedAssetFormDialog", () => {
  it("renders create title with name field", async () => {
    renderWithProviders(
      <FixedAssetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Register fixed asset")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("lists seeded categories in the select", async () => {
    renderWithProviders(
      <FixedAssetFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Register fixed asset");
    expect(screen.getByText("Purchase value")).toBeInTheDocument();
  });
});
