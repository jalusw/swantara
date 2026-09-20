import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Item } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { MoFormDialog } from "../production-order-form-dialog";

function buildProduct(overrides: Partial<Item>): Item {
  return {
    id: 5,
    organizationId: 1,
    name: "Wooden Chair",
    categoryId: null,
    type: "stockable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 0,
    standardCost: 0,
    isPurchasable: true,
    isSellable: true,
    isManufactured: true,
    tracking: "none",
    weight: 0,
    volume: 0,
    hsCode: null,
    descriptionSale: null,
    descriptionPurchase: null,
    active: true,
    createdAt: new Date("2026-01-01T00:00:00Z"),
    updatedAt: new Date("2026-01-01T00:00:00Z"),
    ...overrides,
  };
}

const products = [
  buildProduct({ id: 5, name: "Wooden Chair" }),
  buildProduct({ id: 6, name: "Wooden Table" }),
];

beforeEach(() => {});

describe("MoFormDialog", () => {
  it("renders create title with quantity field", async () => {
    renderWithProviders(
      <MoFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        products={products}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Create Manufacturing Order")).toBeInTheDocument();
    expect(screen.getByLabelText("Quantity to Produce")).toBeInTheDocument();
  });

  it("updates the quantity field", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <MoFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        products={products}
        onSave={vi.fn()}
      />,
    );

    const qty = await screen.findByLabelText("Quantity to Produce");
    await user.clear(qty);
    await user.type(qty, "25");

    expect(qty).toHaveValue(25);
  });
});
