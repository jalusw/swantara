import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Switch } from "@/components/switch";
import { renderWithProviders } from "@/lib/tests";

describe("Switch", () => {
  it("renders with data-slot", () => {
    renderWithProviders(<Switch />);

    expect(screen.getByRole("switch")).toHaveAttribute("data-slot", "switch");
  });

  it("renders as checked", () => {
    renderWithProviders(<Switch checked />);

    expect(screen.getByRole("switch")).toBeChecked();
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Switch className="my-switch" />);

    expect(screen.getByRole("switch")).toHaveClass("my-switch");
  });
});
