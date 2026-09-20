import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import type { StubItem } from "../../../_components/products-data";
import { BomFormDialog } from "../recipe-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const templates: StubItem[] = [
  {
    id: "5",
    name: "Finished Widget",
    categoryId: null,
    type: "stockable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 100,
    standardCost: 60,
    isPurchasable: true,
    isSellable: true,
    isManufactured: true,
    tracking: "none",
    active: true,
  },
  {
    id: "6",
    name: "Raw Material",
    categoryId: null,
    type: "consumable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 10,
    standardCost: 8,
    isPurchasable: true,
    isSellable: false,
    isManufactured: false,
    tracking: "none",
    active: true,
  },
];

const units = [
  {
    id: 1,
    category_id: 1,
    name: "Units",
    factor: 1,
    unit_type: "reference",
    rounding: 0.01,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function seedLists() {
  server.use(
    http.get("*/api/v1/units", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { units } }),
    ),
  );
}

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/recipes", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { recipe: { id: 9 } } });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/recipes/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { recipe: { id: 4 } } });
    }),
  );
}

const editInitial = {
  id: "4",
  itemId: "5",
  code: null,
  type: "manufacture",
  qty: 2,
  unitId: null,
  version: 3,
  active: true,
} as never;

const editLines = [{ id: "l1", componentId: "6", qty: 2, unitId: null, scrapPct: 50 }];

async function selectProduct(user: ReturnType<typeof userEvent.setup>, name: string) {
  await user.click(screen.getByRole("combobox", { name: "Item" }));
  await user.click(await screen.findByRole("option", { name }));
}

beforeEach(() => {
  seedLists();
});

describe("BomFormDialog branches", () => {
  it("shows edit title with prefilled values", async () => {
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        initial={editInitial}
        initialLines={editLines as never}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Edit bill of materials")).toBeInTheDocument();
    expect(screen.getByDisplayValue("3")).toBeInTheDocument();
  });

  it("blocks submit when item and component are missing", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await screen.findByText("New bill of materials");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Select a item.")).toBeInTheDocument();
    expect(await screen.findByText("Select a component.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("blocks submit with zero quantity and version", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await screen.findByText("New bill of materials");
    await selectProduct(user, "Finished Widget");
    const qty = screen.getAllByLabelText("Quantity")[0];
    if (qty === undefined) throw new Error("expected a quantity input");
    await user.clear(qty);
    await user.type(qty, "0");
    const version = screen.getByLabelText("Version");
    await user.clear(version);
    await user.type(version, "0");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Quantity must be greater than zero.")).toBeInTheDocument();
    expect(await screen.findByText("Enter a version number.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("adds and removes component lines", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={vi.fn()}
      />,
    );

    await screen.findByText("New bill of materials");
    await user.click(screen.getByRole("button", { name: "Add component" }));
    expect(screen.getAllByRole("combobox", { name: "Component" })).toHaveLength(2);

    const secondRemove = screen.getAllByRole("button", { name: "Remove line" })[1];
    if (secondRemove === undefined) throw new Error("expected a second remove button");
    await user.click(secondRemove);
    expect(screen.getAllByRole("combobox", { name: "Component" })).toHaveLength(1);
  });

  it("blocks submit when every line was removed", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await screen.findByText("New bill of materials");
    await selectProduct(user, "Finished Widget");
    await user.click(screen.getByRole("button", { name: "Remove line" }));
    expect(screen.queryByRole("combobox", { name: "Component" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).not.toHaveBeenCalled());
  });

  it("creates a recipe with minimal fields", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await screen.findByText("New bill of materials");
    await selectProduct(user, "Finished Widget");
    await user.click(screen.getByRole("combobox", { name: "Component" }));
    await user.click(await screen.findByRole("option", { name: "Raw Material" }));
    await user.type(screen.getByPlaceholderText("Quantity"), "1");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates a recipe with code, unit, scrap, and kit type", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await screen.findByText("New bill of materials");
    await selectProduct(user, "Finished Widget");
    await user.type(screen.getByPlaceholderText("BOM/..."), "BOM-001");
    await user.click(screen.getByRole("combobox", { name: "Type" }));
    await user.click(await screen.findByRole("option", { name: "Kit" }));
    await user.click(screen.getByRole("combobox", { name: "Unit of measure" }));
    await user.click(await screen.findByRole("option", { name: "Units" }));
    await user.click(screen.getByRole("combobox", { name: "Component" }));
    await user.click(await screen.findByRole("option", { name: "Raw Material" }));
    await user.type(screen.getByPlaceholderText("Quantity"), "2");
    await user.type(screen.getByPlaceholderText("Scrap %"), "50");
    await user.click(screen.getByRole("switch", { name: "Active" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Total incl. scrap: 3")).toBeInTheDocument();
    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await screen.findByText("New bill of materials");
    await selectProduct(user, "Finished Widget");
    await user.click(screen.getByRole("combobox", { name: "Component" }));
    await user.click(await screen.findByRole("option", { name: "Raw Material" }));
    await user.type(screen.getByPlaceholderText("Quantity"), "1");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(screen.getByText("New bill of materials")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates a recipe with existing lines", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        initial={editInitial}
        initialLines={editLines as never}
        onSave={onSave}
      />,
    );

    await screen.findByText("Edit bill of materials");
    expect(screen.getByText("Total incl. scrap: 3")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates a recipe without initial lines", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Edit bill of materials");
    expect(screen.queryByRole("combobox", { name: "Component" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Add component" }));
    await user.click(screen.getByRole("combobox", { name: "Component" }));
    await user.click(await screen.findByRole("option", { name: "Raw Material" }));
    await user.type(screen.getByPlaceholderText("Quantity"), "1");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        initial={editInitial}
        initialLines={editLines as never}
        onSave={onSave}
      />,
    );

    await screen.findByText("Edit bill of materials");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(screen.getByText("Edit bill of materials")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={onOpenChange}
        orgId="1"
        templates={templates}
        onSave={vi.fn()}
      />,
    );

    await screen.findByText("New bill of materials");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
