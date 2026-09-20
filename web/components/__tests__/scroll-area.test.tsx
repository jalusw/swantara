import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ScrollArea } from "@/components/scroll-area";
import { renderWithProviders } from "@/lib/tests";

describe("ScrollArea", () => {
  it("renders with data-slot", () => {
    renderWithProviders(
      <ScrollArea>
        <div>Content</div>
      </ScrollArea>,
    );

    expect(screen.getByText("Content").closest("[data-slot='scroll-area']")).toBeInTheDocument();
  });

  it("renders viewport inside scroll area", () => {
    const { container } = renderWithProviders(
      <ScrollArea>
        <div>Content</div>
      </ScrollArea>,
    );

    expect(container.querySelector("[data-slot='scroll-area-viewport']")).toBeInTheDocument();
  });

  it("renders with a custom className", () => {
    renderWithProviders(
      <ScrollArea className="h-64">
        <div>Content</div>
      </ScrollArea>,
    );

    expect(screen.getByText("Content").closest("[data-slot='scroll-area']")).toHaveClass("h-64");
  });
});
