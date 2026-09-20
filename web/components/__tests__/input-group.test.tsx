import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { InputGroup, InputGroupAddon } from "@/components/input-group";
import { renderWithProviders } from "@/lib/tests";

describe("InputGroup", () => {
  it("renders with data-slot and role", () => {
    renderWithProviders(
      <InputGroup>
        <input />
      </InputGroup>,
    );

    expect(screen.getByRole("group")).toHaveAttribute("data-slot", "input-group");
  });

  it("renders addon with data-slot", () => {
    renderWithProviders(
      <InputGroup>
        <InputGroupAddon data-slot="input-group-addon">$</InputGroupAddon>
        <input />
      </InputGroup>,
    );

    expect(screen.getByText("$")).toHaveAttribute("data-slot", "input-group-addon");
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(
      <InputGroup className="my-group">
        <input />
      </InputGroup>,
    );

    expect(container.firstElementChild).toHaveClass("my-group");
  });
});
