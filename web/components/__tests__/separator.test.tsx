import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Separator } from "@/components/separator";
import { renderWithProviders } from "@/lib/tests";

describe("Separator", () => {
  it("renders horizontal by default", () => {
    renderWithProviders(<Separator />);

    expect(screen.getByRole("separator")).toHaveAttribute("data-slot", "separator");
    expect(screen.getByRole("separator")).toHaveAttribute("data-orientation", "horizontal");
  });

  it("renders vertical orientation", () => {
    renderWithProviders(<Separator orientation="vertical" />);

    expect(screen.getByRole("separator")).toHaveAttribute("data-orientation", "vertical");
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Separator className="my-separator" />);

    expect(screen.getByRole("separator")).toHaveClass("my-separator");
  });
});
