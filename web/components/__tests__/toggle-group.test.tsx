import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ToggleGroup, ToggleGroupItem } from "@/components/toggle-group";
import { renderWithProviders } from "@/lib/tests";

describe("ToggleGroup", () => {
  it("renders group with data-slot", () => {
    renderWithProviders(
      <ToggleGroup>
        <ToggleGroupItem value="a">A</ToggleGroupItem>
        <ToggleGroupItem value="b">B</ToggleGroupItem>
      </ToggleGroup>,
    );

    expect(screen.getByRole("group")).toHaveAttribute("data-slot", "toggle-group");
  });

  it("renders items with data-slot", () => {
    renderWithProviders(
      <ToggleGroup>
        <ToggleGroupItem value="a">Option A</ToggleGroupItem>
      </ToggleGroup>,
    );

    expect(screen.getByRole("button", { name: "Option A" })).toHaveAttribute(
      "data-slot",
      "toggle-group-item",
    );
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(
      <ToggleGroup className="my-group">
        <ToggleGroupItem value="a">A</ToggleGroupItem>
      </ToggleGroup>,
    );

    expect(container.firstElementChild).toHaveClass("my-group");
  });
});
