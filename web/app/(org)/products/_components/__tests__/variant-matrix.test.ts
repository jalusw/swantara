import { describe, expect, it } from "vitest";
import type { AttributeOption } from "../attribute-editor";
import { buildMatrix, generateSku } from "../variant-matrix-preview";

describe("buildMatrix", () => {
  it("returns a single empty combination when no attributes are provided", () => {
    expect(buildMatrix([])).toEqual([{}]);
  });

  it("filters out attributes with empty names", () => {
    const attributes: AttributeOption[] = [
      { name: "", values: ["Red", "Blue"] },
      { name: "Size", values: ["S", "M"] },
    ];
    const matrix = buildMatrix(attributes);
    expect(matrix).toEqual([{ Size: "S" }, { Size: "M" }]);
  });

  it("filters out attributes with no values", () => {
    const attributes: AttributeOption[] = [
      { name: "Color", values: [] },
      { name: "Size", values: ["S"] },
    ];
    expect(buildMatrix(attributes)).toEqual([{ Size: "S" }]);
  });

  it("computes the cartesian item of two attributes", () => {
    const attributes: AttributeOption[] = [
      { name: "Color", values: ["Red", "Blue"] },
      { name: "Size", values: ["S", "M", "L"] },
    ];
    const matrix = buildMatrix(attributes);
    expect(matrix).toEqual([
      { Color: "Red", Size: "S" },
      { Color: "Red", Size: "M" },
      { Color: "Red", Size: "L" },
      { Color: "Blue", Size: "S" },
      { Color: "Blue", Size: "M" },
      { Color: "Blue", Size: "L" },
    ]);
  });

  it("computes the cartesian item of three attributes", () => {
    const attributes: AttributeOption[] = [
      { name: "Color", values: ["Red"] },
      { name: "Size", values: ["S", "M"] },
      { name: "Finish", values: ["Matte", "Gloss"] },
    ];
    const matrix = buildMatrix(attributes);
    expect(matrix).toHaveLength(4);
    expect(matrix[0]).toEqual({
      Color: "Red",
      Size: "S",
      Finish: "Matte",
    });
    expect(matrix[3]).toEqual({
      Color: "Red",
      Size: "M",
      Finish: "Gloss",
    });
  });
});

describe("generateSku", () => {
  it("generates a SKU from template name and attributes", () => {
    const sku = generateSku("Classic Tee", { Color: "Red", Size: "S" });
    expect(sku).toBe("CLATEE-RED-S");
  });

  it("returns just the base when no attributes are provided", () => {
    const sku = generateSku("Standing Desk", {});
    expect(sku).toBe("STADES");
  });

  it("strips special characters from the template name", () => {
    const sku = generateSku("T-Shirt (V-Neck)", {});
    expect(sku).toBe("TSHVNE");
  });

  it("truncates the base to 6 characters", () => {
    const sku = generateSku("Long Template Name", {});
    expect(sku).toBe("LONTEM");
  });

  it("uppercases the entire SKU", () => {
    const sku = generateSku("my item", { color: "red" });
    expect(sku).toBe("MYITE-RED");
  });
});
