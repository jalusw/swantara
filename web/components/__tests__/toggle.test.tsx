import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Toggle } from "@/components/toggle";
import { renderWithProviders } from "@/lib/tests";

describe("Toggle", () => {
  it("renders a toggle button with data-slot", () => {
    renderWithProviders(<Toggle>Toggle</Toggle>);

    expect(screen.getByRole("button", { name: "Toggle" })).toHaveAttribute("data-slot", "toggle");
  });

  it("renders as pressed", () => {
    renderWithProviders(<Toggle pressed>Toggle</Toggle>);

    expect(screen.getByRole("button", { name: "Toggle" })).toHaveAttribute("aria-pressed", "true");
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Toggle className="my-toggle">Toggle</Toggle>);

    expect(screen.getByRole("button", { name: "Toggle" })).toHaveClass("my-toggle");
  });
});
