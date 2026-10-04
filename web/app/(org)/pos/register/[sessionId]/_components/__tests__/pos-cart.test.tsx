import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PosCart } from "../pos-cart";
import type { CartLine } from "../pos-register-section";

const lines: CartLine[] = [
  {
    key: "k1",
    itemId: 1,
    productName: "Arabica Beans",
    qty: 2,
    unitPrice: 50000,
    discountPct: 0,
    taxIds: [],
  },
  {
    key: "k2",
    itemId: 2,
    productName: "Paper Cups",
    qty: 1,
    unitPrice: 15000,
    discountPct: 10,
    taxIds: [],
  },
];

describe("PosCart", () => {
  it("renders empty state when there are no lines", () => {
    renderWithProviders(
      <PosCart
        lines={[]}
        subtotal={0}
        onUpdateLine={() => {}}
        onRemoveLine={() => {}}
        onClear={() => {}}
      />,
    );

    expect(screen.getByText("Keranjang Kosong")).toBeInTheDocument();
  });

  it("renders lines with quantities and subtotal", () => {
    renderWithProviders(
      <PosCart
        lines={lines}
        subtotal={114000}
        onUpdateLine={() => {}}
        onRemoveLine={() => {}}
        onClear={() => {}}
      />,
    );

    expect(screen.getByText("Arabica Beans")).toBeInTheDocument();
    expect(screen.getByText("Paper Cups")).toBeInTheDocument();
    expect(screen.getByText("2 item")).toBeInTheDocument();
    expect(screen.getByText("Subtotal")).toBeInTheDocument();
  });

  it("clears the cart from the header button", async () => {
    const user = userEvent.setup();
    const onClear = vi.fn();
    renderWithProviders(
      <PosCart
        lines={lines}
        subtotal={114000}
        onUpdateLine={() => {}}
        onRemoveLine={() => {}}
        onClear={onClear}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Clear" }));

    expect(onClear).toHaveBeenCalledTimes(1);
  });
});
