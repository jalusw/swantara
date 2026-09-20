import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PosPayment } from "../pos-payment";

function renderPayment(onSubmit: (payments: Array<{ method: string; amount: number }>) => void) {
  renderWithProviders(
    <PosPayment open onOpenChange={() => {}} total={120000} onSubmit={onSubmit} />,
  );
}

describe("PosPayment", () => {
  it("renders the amount due", async () => {
    renderPayment(() => {});

    expect(await screen.findByText("Total due")).toBeInTheDocument();
    expect(screen.getByText("Remaining")).toBeInTheDocument();
  });

  it("submits the balanced payment on confirm", async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    renderPayment(onSubmit);

    await screen.findByText("Total due");
    await user.click(screen.getByRole("button", { name: "Confirm" }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith([{ method: "cash", amount: 120000 }]);
  });
});
