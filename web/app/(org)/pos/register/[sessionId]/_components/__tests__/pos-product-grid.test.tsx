import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { Item } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { PosProductGrid } from "../pos-product-grid";

function buildProduct(overrides: Partial<Item>): Item {
  return {
    id: 1,
    organizationId: 1,
    name: "Arabica Beans",
    categoryId: null,
    type: "stockable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 50000,
    standardCost: 30000,
    isPurchasable: true,
    isSellable: true,
    isManufactured: false,
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
  buildProduct({ id: 1, name: "Arabica Beans", type: "stockable" }),
  buildProduct({ id: 2, name: "Paper Cups", type: "consumable", listPrice: 15000 }),
];

describe("PosProductGrid", () => {
  it("renders products with types and prices", async () => {
    renderWithProviders(
      <PosProductGrid
        products={products}
        onAddItem={() => {}}
        searchPlaceholder="Search products..."
      />,
    );

    expect(await screen.findByText("Arabica Beans")).toBeInTheDocument();
    expect(screen.getByText("Paper Cups")).toBeInTheDocument();
    expect(screen.getByText("stockable")).toBeInTheDocument();
  });

  it("filters products through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PosProductGrid
        products={products}
        onAddItem={() => {}}
        searchPlaceholder="Search products..."
      />,
    );

    await screen.findByText("Arabica Beans");
    await user.type(screen.getByPlaceholderText("Search products..."), "cups");

    expect(await screen.findByText("Paper Cups")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Arabica Beans")).not.toBeInTheDocument());
  });

  it("adds a item when its card is clicked", async () => {
    const user = userEvent.setup();
    const onAddItem = vi.fn();
    renderWithProviders(
      <PosProductGrid
        products={products}
        onAddItem={onAddItem}
        searchPlaceholder="Search products..."
      />,
    );

    await user.click(await screen.findByText("Paper Cups"));

    expect(onAddItem).toHaveBeenCalledTimes(1);
    expect(onAddItem).toHaveBeenCalledWith(expect.objectContaining({ id: 2 }));
  });
});
