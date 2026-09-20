import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import type { ItemVariant } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { ItemVariants } from "../product-variants";

const variants: ItemVariant[] = [
  {
    id: 21,
    templateId: 5,
    sku: "TOTE-RED",
    barcode: "899000000021",
    attributeJson: { Color: "Red" },
    extraCost: 0,
    active: true,
    createdAt: new Date("2026-01-01T00:00:00Z"),
    updatedAt: new Date("2026-01-01T00:00:00Z"),
  },
  {
    id: 22,
    templateId: 5,
    sku: "TOTE-BLUE",
    barcode: null,
    attributeJson: { Color: "Blue" },
    extraCost: 2,
    active: true,
    createdAt: new Date("2026-01-01T00:00:00Z"),
    updatedAt: new Date("2026-01-01T00:00:00Z"),
  },
];

describe("ItemVariants", () => {
  it("renders variant SKUs with attribute badges", () => {
    renderWithProviders(<ItemVariants variants={variants} />);

    expect(screen.getByText("TOTE-RED")).toBeInTheDocument();
    expect(screen.getByText("TOTE-BLUE")).toBeInTheDocument();
    expect(screen.getByText("Color: Red")).toBeInTheDocument();
  });

  it("filters variants through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ItemVariants variants={variants} />);

    await user.type(screen.getByPlaceholderText("SKU…"), "BLUE");

    expect(await screen.findByText("TOTE-BLUE")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("TOTE-RED")).not.toBeInTheDocument());
  });

  it("shows an empty message when no variants exist", () => {
    renderWithProviders(<ItemVariants variants={[]} />);

    expect(screen.getByText("This item has no variants yet.")).toBeInTheDocument();
  });
});
