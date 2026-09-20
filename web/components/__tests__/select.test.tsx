import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/select";
import { renderWithProviders } from "@/lib/tests";

describe("Select", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <Select>
        <SelectTrigger>Choose</SelectTrigger>
      </Select>,
    );

    expect(screen.getByRole("combobox")).toHaveAttribute("data-slot", "select-trigger");
  });

  it("renders items when open", () => {
    renderWithProviders(
      <Select open>
        <SelectTrigger>Choose</SelectTrigger>
        <SelectContent>
          <SelectItem value="a">Apple</SelectItem>
          <SelectItem value="b">Banana</SelectItem>
        </SelectContent>
      </Select>,
    );

    expect(screen.getByText("Apple")).toBeInTheDocument();
    expect(screen.getByText("Banana")).toBeInTheDocument();
  });

  it("passes through a custom className on trigger", () => {
    renderWithProviders(
      <Select>
        <SelectTrigger className="my-select">Choose</SelectTrigger>
      </Select>,
    );

    expect(screen.getByRole("combobox")).toHaveClass("my-select");
  });
});
