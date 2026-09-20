import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Label } from "@/components/label";
import { renderWithProviders } from "@/lib/tests";

describe("Label", () => {
  it("renders a label element with data-slot", () => {
    renderWithProviders(<Label htmlFor="email">Email</Label>);

    const label = screen.getByText("Email");
    expect(label).toHaveAttribute("data-slot", "label");
    expect(label).toHaveAttribute("for", "email");
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Label className="my-label">Name</Label>);

    expect(screen.getByText("Name")).toHaveClass("my-label");
  });
});
