import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Barcode, encodeCode128B } from "@/components/barcode";
import { renderWithProviders } from "@/lib/tests";

describe("encodeCode128B", () => {
  it("produces a deterministic bit stream that starts with Start-B", () => {
    expect(encodeCode128B("5")).toBe(
      "11010010000" + "11011100100" + "11001110100" + "1100011101011",
    );
  });
});

describe("Barcode", () => {
  it("renders an accessible barcode for the value", () => {
    renderWithProviders(<Barcode value="SW-1042" />);

    expect(screen.getByRole("img", { name: "Barcode SW-1042" })).toBeInTheDocument();
    expect(screen.getByText("SW-1042")).toBeInTheDocument();
  });

  it("draws bars in the svg", () => {
    const { container } = renderWithProviders(<Barcode value="SW-1042" />);

    expect(container.querySelectorAll("[data-slot='barcode'] svg rect").length).toBeGreaterThan(0);
  });

  it("falls back when the value cannot be encoded", () => {
    renderWithProviders(<Barcode value="" />);

    expect(screen.getByText("Empty barcode")).toBeInTheDocument();
  });
});
