import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import type { AttributeRow } from "../attribute-editor";
import { VariantMatrixPreview } from "../variant-matrix-preview";

describe("VariantMatrixPreview", () => {
  it("shows a message when no attributes are defined", () => {
    renderWithProviders(<VariantMatrixPreview templateName="Barang" attributes={[]} />);
    expect(
      screen.getByText("Tambahkan atribut dengan nilai untuk melihat pratinjau varian."),
    ).toBeInTheDocument();
  });

  it("renders the cartesian item of attributes as rows", () => {
    const attributes: AttributeRow[] = [
      { id: "a1", name: "Color", values: ["Red", "Blue"] },
      { id: "a2", name: "Size", values: ["S", "M"] },
    ];
    renderWithProviders(<VariantMatrixPreview templateName="Tee" attributes={attributes} />);
    expect(screen.getByText("4 varian akan dibuat.")).toBeInTheDocument();
    expect(screen.getByText("Color: Red · Size: S")).toBeInTheDocument();
    expect(screen.getByText("Color: Blue · Size: M")).toBeInTheDocument();
  });

  it("surfaces duplicate SKUs with a destructive alert", () => {
    const attributes: AttributeRow[] = [{ id: "a1", name: "Color", values: ["Red", "RED"] }];
    renderWithProviders(<VariantMatrixPreview templateName="Tee" attributes={attributes} />);
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByText(/SKU duplikat ditemukan/)).toBeInTheDocument();
  });

  it("does not show a duplicate alert when SKUs are unique", () => {
    const attributes: AttributeRow[] = [{ id: "a1", name: "Color", values: ["Red", "Blue"] }];
    renderWithProviders(<VariantMatrixPreview templateName="Tee" attributes={attributes} />);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
