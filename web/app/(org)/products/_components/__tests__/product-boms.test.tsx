import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import type { Recipe } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { ProductBoms } from "../product-boms";

const recipes: Recipe[] = [
  {
    id: 31,
    organizationId: 1,
    itemId: 5,
    code: "BOM-TOTE-001",
    qty: 1,
    unitId: null,
    type: "manufacture",
    version: 2,
    active: true,
    createdAt: new Date("2026-01-01T00:00:00Z"),
    updatedAt: new Date("2026-01-01T00:00:00Z"),
  },
  {
    id: 32,
    organizationId: 1,
    itemId: 6,
    code: "BOM-MUG-001",
    qty: 24,
    unitId: null,
    type: "kit",
    version: 1,
    active: true,
    createdAt: new Date("2026-01-01T00:00:00Z"),
    updatedAt: new Date("2026-01-01T00:00:00Z"),
  },
];

describe("ProductBoms", () => {
  it("renders recipe codes with type and version", () => {
    renderWithProviders(<ProductBoms recipes={recipes} />);

    expect(screen.getByText("BOM-TOTE-001")).toBeInTheDocument();
    expect(screen.getByText("BOM-MUG-001")).toBeInTheDocument();
    expect(screen.getByText("Manufaktur")).toBeInTheDocument();
    expect(screen.getByText("v2")).toBeInTheDocument();
  });

  it("filters recipes through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProductBoms recipes={recipes} />);

    await user.type(screen.getByPlaceholderText("Cari resep"), "MUG");

    expect(await screen.findByText("BOM-MUG-001")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("BOM-TOTE-001")).not.toBeInTheDocument());
  });

  it("shows an empty message when no recipes exist", () => {
    renderWithProviders(<ProductBoms recipes={[]} />);

    expect(screen.getByText("Belum ada resep")).toBeInTheDocument();
  });
});
