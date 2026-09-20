import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Button } from "@/components/button";
import { renderWithProviders } from "@/lib/tests";

describe("Button", () => {
  it("renders a button with the default variant and size", () => {
    renderWithProviders(<Button>Click me</Button>);

    expect(screen.getByRole("button", { name: "Click me" })).toHaveAttribute("data-slot", "button");
  });

  it("applies variant and size classes", () => {
    const { container } = renderWithProviders(
      <Button variant="destructive" size="sm">
        Delete
      </Button>,
    );

    expect(container.firstElementChild).toHaveClass("text-destructive", "h-11");
  });

  it("renders as a child element when asChild is set", () => {
    renderWithProviders(
      <Button asChild>
        <a href="/dashboard">Go</a>
      </Button>,
    );

    const link = screen.getByRole("link", { name: "Go" });
    expect(link).toHaveAttribute("href", "/dashboard");
    expect(link).toHaveAttribute("data-slot", "button");
  });

  describe("regression: link buttons", () => {
    it("should render anchor as link preserving navigation semantics", () => {
      renderWithProviders(
        <Button asChild variant="outline">
          <a href="/dashboard">Dashboard</a>
        </Button>,
      );

      const link = screen.getByRole("link", { name: "Dashboard" });
      expect(link).toHaveAttribute("href", "/dashboard");
      expect(link).toHaveAttribute("data-slot", "button");
    });
  });
});
