import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Textarea } from "@/components/textarea";
import { renderWithProviders } from "@/lib/tests";

describe("Textarea", () => {
  it("renders a textarea with data-slot", () => {
    renderWithProviders(<Textarea />);

    expect(screen.getByRole("textbox")).toHaveAttribute("data-slot", "textarea");
  });

  it("renders with a placeholder", () => {
    renderWithProviders(<Textarea placeholder="Enter text" />);

    expect(screen.getByPlaceholderText("Enter text")).toBeInTheDocument();
  });

  it("renders as disabled", () => {
    renderWithProviders(<Textarea disabled />);

    expect(screen.getByRole("textbox")).toBeDisabled();
  });
});
