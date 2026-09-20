import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PosReceipt } from "../pos-receipt";

function renderReceipt(onNewOrder: () => void) {
  renderWithProviders(
    <PosReceipt
      orderId={9}
      orderName="POS-0009"
      amountTotal={120000}
      payments={[{ method: "cash", amount: 120000 }]}
      onNewOrder={onNewOrder}
    />,
  );
}

describe("PosReceipt", () => {
  it("renders the order receipt with payments", async () => {
    renderReceipt(() => {});

    expect(await screen.findByText("POS-0009")).toBeInTheDocument();
    expect(screen.getByText("Receipt")).toBeInTheDocument();
    expect(screen.getByText("cash")).toBeInTheDocument();
  });

  it("starts a new order from the receipt", async () => {
    const onNewOrder = vi.fn();
    const user = userEvent.setup();
    renderReceipt(onNewOrder);

    await screen.findByText("POS-0009");
    await user.click(screen.getByRole("button", { name: "New order" }));

    expect(onNewOrder).toHaveBeenCalledTimes(1);
  });
});
