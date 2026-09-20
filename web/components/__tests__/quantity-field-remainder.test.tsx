import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { QuantityField } from "@/components/quantity-field";
import { renderWithProviders } from "@/lib/tests";

describe("QuantityField remainder", () => {
  it("ignores stepper clicks when disabled", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<QuantityField value={2} onValueChange={onChange} disabled />);

    await user.click(screen.getByRole("button", { name: "Decrease quantity" }));
    await user.click(screen.getByRole("button", { name: "Increase quantity" }));

    expect(onChange).not.toHaveBeenCalled();
  });

  it("clears the value when input is blanked and blurred", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<QuantityField value={5} onValueChange={onChange} />);

    const input = screen.getByRole("textbox", { name: "Quantity" });
    await user.clear(input);
    await user.tab();

    expect(onChange).toHaveBeenLastCalledWith(null);
  });

  it("announces clamping when blur exceeds the maximum", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<QuantityField value={1} max={5} onValueChange={onChange} />);

    const input = screen.getByRole("textbox", { name: "Quantity" });
    await user.clear(input);
    await user.type(input, "99");
    await user.tab();

    expect(onChange).toHaveBeenLastCalledWith(5);
    expect(screen.getByText("Value adjusted to 5.")).toBeInTheDocument();
  });

  it("clamps typed input below the minimum on change", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<QuantityField value={5} min={2} onValueChange={onChange} />);

    const input = screen.getByRole("textbox", { name: "Quantity" });
    await user.clear(input);
    await user.type(input, "1");

    expect(onChange).toHaveBeenLastCalledWith(2);
  });

  it("hides the unit select when units list is empty", () => {
    renderWithProviders(
      <QuantityField value={1} onValueChange={vi.fn()} units={[]} unit="" onUnitChange={vi.fn()} />,
    );

    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("applies id, placeholder and custom aria-label", () => {
    renderWithProviders(
      <QuantityField
        value={null}
        onValueChange={vi.fn()}
        id="qty-input"
        placeholder="Enter qty"
        aria-label="Custom quantity"
      />,
    );

    const input = screen.getByRole("textbox", { name: "Custom quantity" });
    expect(input).toHaveAttribute("id", "qty-input");
    expect(input).toHaveAttribute("placeholder", "Enter qty");
  });
});
