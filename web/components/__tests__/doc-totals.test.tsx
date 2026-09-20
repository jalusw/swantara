import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DocTotals } from "@/components/doc-totals";
import { renderWithProviders } from "@/lib/tests";

describe("DocTotals", () => {
  it("renders subtotal and computes the total", () => {
    renderWithProviders(<DocTotals subtotal={1900} />);

    expect(screen.getByText("Subtotal")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getAllByText("$1,900.00")).toHaveLength(2);
  });

  it("renders discount, taxes and shipping with the resulting balance", () => {
    renderWithProviders(
      <DocTotals
        subtotal={1900}
        discount={100}
        taxes={[{ label: "VAT 20%", amount: 380 }]}
        shipping={140}
      />,
    );

    expect(screen.getByText("Discount")).toBeInTheDocument();
    expect(screen.getByText("-$100.00")).toBeInTheDocument();
    expect(screen.getByText("VAT 20%")).toBeInTheDocument();
    expect(screen.getByText("$380.00")).toBeInTheDocument();
    expect(screen.getByText("Shipping")).toBeInTheDocument();
    expect(screen.getByText("$2,320.00")).toBeInTheDocument();
  });

  it("omits optional rows when they are zero", () => {
    renderWithProviders(<DocTotals subtotal={500} discount={0} />);

    expect(screen.queryByText("Discount")).not.toBeInTheDocument();
    expect(screen.queryByText("Shipping")).not.toBeInTheDocument();
  });
});
