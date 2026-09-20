import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { QuantityField } from "@/components/quantity-field";
import { renderWithProviders } from "@/lib/tests";

describe("QuantityField", () => {
  it("increments and decrements with the steppers", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<QuantityField value={2} onValueChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Increase quantity" }));
    expect(onChange).toHaveBeenLastCalledWith(3);

    await user.click(screen.getByRole("button", { name: "Decrease quantity" }));
    expect(onChange).toHaveBeenLastCalledWith(1);
  });

  it("starts from zero when the value is empty", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<QuantityField value={null} onValueChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Increase quantity" }));
    expect(onChange).toHaveBeenLastCalledWith(1);
  });

  it("disables steppers at the min and max bounds", () => {
    const { rerender } = renderWithProviders(
      <QuantityField value={0} min={0} max={5} onValueChange={vi.fn()} />,
    );
    expect(screen.getByRole("button", { name: "Decrease quantity" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Increase quantity" })).not.toBeDisabled();

    rerender(<QuantityField value={5} min={0} max={5} onValueChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Increase quantity" })).toBeDisabled();
  });

  it("passes a unit selection through when units are provided", async () => {
    const user = userEvent.setup();
    const onUnitChange = vi.fn();
    renderWithProviders(
      <QuantityField
        value={1}
        onValueChange={vi.fn()}
        unit="kg"
        onUnitChange={onUnitChange}
        units={[
          { value: "kg", label: "kg" },
          { value: "box", label: "box" },
        ]}
      />,
    );

    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByRole("option", { name: "box" }));
    expect(onUnitChange).toHaveBeenCalledWith("box");
  });
});
