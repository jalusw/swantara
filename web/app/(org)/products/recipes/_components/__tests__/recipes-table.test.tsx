import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import type { StubBom, StubItem } from "../../../_components/products-data";
import { BomsTable } from "../recipes-table";

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
];

const recipes: StubBom[] = [
  {
    id: "1",
    itemId: "5",
    code: "BOM-001",
    qty: 1,
    unitId: null,
    type: "manufacture",
    version: 1,
    active: true,
    lines: [{ id: "l1", componentId: "6", qty: 2, unitId: null, scrapPct: 0 }],
  },
  {
    id: "2",
    itemId: "5",
    code: "BOM-002",
    qty: 5,
    unitId: null,
    type: "kit",
    version: 2,
    active: false,
    lines: [],
  },
];

describe("BomsTable", () => {
  it("renders recipe codes with resolved item names", () => {
    renderWithProviders(<BomsTable recipes={recipes} templates={templates} />);

    expect(screen.getByText("BOM-001")).toBeInTheDocument();
    expect(screen.getByText("BOM-002")).toBeInTheDocument();
    expect(screen.getAllByText("Finished Widget").length).toBeGreaterThan(0);
  });

  it("filters rows by search text", async () => {
    const user = userEvent.setup();
    renderWithProviders(<BomsTable recipes={recipes} templates={templates} />);

    await user.type(screen.getByPlaceholderText("Cari resep"), "BOM-002");

    await waitFor(() => expect(screen.queryByText("BOM-001")).not.toBeInTheDocument());
    expect(screen.getAllByText("BOM-002").length).toBeGreaterThan(0);
  });
});
