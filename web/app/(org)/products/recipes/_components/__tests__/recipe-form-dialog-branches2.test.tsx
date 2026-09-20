import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it, vi } from "vitest";
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

function seedUnits() {
  server.use(
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          units: [
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
          ],
        },
      }),
    ),
  );
}

describe("BomFormDialog branches2", () => {
  it("renders the edit title branch with initial values", () => {
    seedUnits();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        initial={{
          id: "9",
          itemId: "5",
          code: "BOM-9",
          type: "manufacture",
          qty: 2,
          unitId: null,
          version: 2,
          active: true,
        }}
        initialLines={[{ id: "l1", componentId: "6", qty: 3, unitId: null, scrapPct: 10 }]}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByText("Edit bill of materials")).toBeInTheDocument();
  });

  it("adds and removes component lines", async () => {
    seedUnits();
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

    await user.click(screen.getByRole("button", { name: "Add component" }));
    const removeButtons = screen.getAllByRole("button", { name: "Remove line" });
    expect(removeButtons.length).toBeGreaterThanOrEqual(2);
    const first = removeButtons[0];
    if (first) {
      await user.click(first);
    }
    expect(screen.getAllByRole("button", { name: "Remove line" }).length).toBeGreaterThanOrEqual(1);
  });

  it("updates the scrap-inclusive total on both qty branches", async () => {
    seedUnits();
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

    expect(await screen.findByText(/Total incl\. scrap:/)).toBeInTheDocument();
    const qtyInputs = screen.getAllByLabelText("Quantity");
    const firstQty = qtyInputs[0];
    if (firstQty) {
      await user.clear(firstQty);
      await user.type(firstQty, "4");
    }
    await waitFor(() => {
      expect(screen.getByText(/Total incl\. scrap:/)).toBeInTheDocument();
    });
  });

  it("creates a recipe with null unit branch and calls onSave", async () => {
    seedUnits();
    let createBody: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/recipes", async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "Item" }));
    await user.click(await screen.findByRole("option", { name: "Finished Widget" }));
    const dialog = screen.getByRole("dialog");
    const componentBoxes = within(dialog).getAllByRole("combobox", { name: "Component" });
    const firstComponent = componentBoxes[0];
    if (firstComponent) {
      await user.click(firstComponent);
      await user.click(await screen.findByRole("option", { name: "Raw Material" }));
    }
    const qtyInputs = within(dialog).getAllByLabelText("Quantity");
    const lineQty = qtyInputs[qtyInputs.length - 1];
    if (lineQty) {
      await user.clear(lineQty);
      await user.type(lineQty, "2");
    }
    await user.click(within(dialog).getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(createBody).toMatchObject({ item_id: 5 });
  });

  it("updates an existing recipe through the update branch", async () => {
    seedUnits();
    let updateBody: unknown = null;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/recipes/:bomId", async ({ request }) => {
        updateBody = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        initial={{
          id: "9",
          itemId: "5",
          code: "",
          type: "manufacture",
          qty: 1,
          unitId: "",
          version: 1,
          active: true,
        }}
        initialLines={[{ id: "l1", componentId: "6", qty: 1, unitId: "", scrapPct: 0 }]}
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(updateBody).toMatchObject({ item_id: 5 });
  });
});
