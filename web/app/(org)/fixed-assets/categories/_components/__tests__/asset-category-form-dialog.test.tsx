import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AssetCategoryFormDialog } from "../asset-category-form-dialog";

function useLocalAccounts() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts: [] } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/asset-categories", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          asset_category: {
            id: 3,
            organization_id: 1,
            name: "IT Equipment",
            asset_account_id: null,
            depreciation_account_id: null,
            expense_account_id: null,
            gain_account_id: null,
            loss_account_id: null,
            method: "linear",
            method_number: null,
            method_period: "year",
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalAccounts();
});

describe("AssetCategoryFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Buat kategori aset")).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("creates a new asset category", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Nama"), "IT Equipment");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("3"));
  });
});
