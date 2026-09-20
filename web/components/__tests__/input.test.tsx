import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Input } from "@/components/input";
import { renderWithProviders } from "@/lib/tests";

describe("Input", () => {
  it("renders an input with data-slot", () => {
    renderWithProviders(<Input />);

    expect(screen.getByRole("textbox")).toHaveAttribute("data-slot", "input");
  });

  it("renders with a placeholder", () => {
    renderWithProviders(<Input placeholder="Enter text" />);

    expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument();
  });

  it("renders as disabled", () => {
    renderWithProviders(<Input disabled />);

    expect(screen.getByRole("textbox")).toBeDisabled();
  });
});
