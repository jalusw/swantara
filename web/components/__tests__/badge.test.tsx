import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Badge } from "@/components/badge";
import { renderWithProviders } from "@/lib/tests";

describe("Badge", () => {
  it("renders its content with the default variant", () => {
    renderWithProviders(<Badge>New</Badge>);

    expect(screen.getByText("New")).toBeInTheDocument();
  });

  it("applies a variant class", () => {
    const { container } = renderWithProviders(<Badge variant="destructive">Error</Badge>);

    expect(container.firstElementChild).toHaveClass("bg-destructive/10", "text-destructive");
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(<Badge className="my-badge">Tag</Badge>);

    expect(container.firstElementChild).toHaveClass("my-badge");
  });
});
