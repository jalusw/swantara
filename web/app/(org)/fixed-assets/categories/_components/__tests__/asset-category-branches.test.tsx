import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AssetCategoryFormDialog } from "../asset-category-form-dialog";

const accounts = [
  { id: 41, code: "1500", name: "Equipment" },
  { id: 42, code: "1501", name: "Accumulated Depreciation" },
];

function seedLists() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts } }),
    ),
  );
}

function seedSave(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/asset-categories", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
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
      });
    }),
  );
}

const editInitial = {
  id: 7,
  name: "Vehicles",
  method: "declining",
  methodNumber: null,
  methodPeriod: "month",
  assetAccountId: null,
  depreciationAccountId: null,
  expenseAccountId: null,
  gainAccountId: null,
  lossAccountId: null,
} as never;

const editInitialFilled = {
  id: 8,
  name: "IT Equipment",
  method: "linear",
  methodNumber: 4,
  methodPeriod: "year",
  assetAccountId: 41,
  depreciationAccountId: 42,
  expenseAccountId: 41,
  gainAccountId: 42,
  lossAccountId: 41,
} as never;

const editInitialFallback = {
  id: 9,
  name: "Furniture",
  method: null,
  methodNumber: null,
  methodPeriod: null,
  assetAccountId: null,
  depreciationAccountId: null,
  expenseAccountId: null,
  gainAccountId: null,
  lossAccountId: null,
} as never;

beforeEach(() => {
  seedLists();
});

describe("AssetCategoryFormDialog branches", () => {
  it("shows edit title with prefilled values", async () => {
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Edit asset category")).toBeInTheDocument();
    expect(screen.getByDisplayValue("IT Equipment")).toBeInTheDocument();
    expect(screen.getByDisplayValue("4")).toBeInTheDocument();
  });

  it("falls back to linear and year when initial lacks them", async () => {
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFallback}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Edit asset category")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Furniture")).toBeInTheDocument();
  });

  it("blocks submit when name is empty", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        onSave={onSave}
      />,
    );

    await screen.findByText("Create asset category");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Name is required")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates a category with only a name", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Name"), "IT Equipment");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("3"));
  });

  it("creates a category with declining method and accounts", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Name"), "Vehicles");
    await user.click(screen.getByRole("combobox", { name: "Depreciation method" }));
    await user.click(
      await screen.findByRole("option", {
        name: "Declining",
      }),
    );
    await user.type(screen.getByLabelText("Method number"), "5");
    await user.click(screen.getByRole("combobox", { name: "Period" }));
    await user.click(
      await screen.findByRole("option", {
        name: "Month",
      }),
    );
    await user.click(screen.getByRole("combobox", { name: "Asset account" }));
    await user.click(await screen.findByRole("option", { name: "1500 Equipment" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("3"));
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave(500);
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={null}
        onSave={onSave}
      />,
    );

    await user.type(await screen.findByLabelText("Name"), "IT Equipment");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(screen.getByLabelText("Name")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates a category with empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Edit asset category");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("7"));
  });

  it("updates a category with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={onSave}
      />,
    );

    await screen.findByText("Edit asset category");
    await user.clear(screen.getByDisplayValue("4"));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("8"));
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedSave(500);
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Edit asset category");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(screen.getByText("Edit asset category")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <AssetCategoryFormDialog
        open={true}
        onOpenChange={onOpenChange}
        orgId="1"
        initial={null}
        onSave={vi.fn()}
      />,
    );

    await screen.findByText("Create asset category");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
