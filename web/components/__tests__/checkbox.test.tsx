import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Checkbox } from "@/components/checkbox";
import { renderWithProviders } from "@/lib/tests";

describe("Checkbox", () => {
  it("renders a checkbox with data-slot", () => {
    renderWithProviders(<Checkbox />);

    expect(screen.getByRole("checkbox")).toHaveAttribute("data-slot", "checkbox");
  });

  it("can be rendered as checked", () => {
    renderWithProviders(<Checkbox checked />);

    expect(screen.getByRole("checkbox")).toBeChecked();
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Checkbox className="my-checkbox" />);

    expect(screen.getByRole("checkbox")).toHaveClass("my-checkbox");
  });
});
