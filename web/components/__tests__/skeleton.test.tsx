import { describe, expect, it } from "vitest";
import { Skeleton } from "@/components/skeleton";
import { renderWithProviders } from "@/lib/tests";

describe("Skeleton", () => {
  it("renders with data-slot", () => {
    const { container } = renderWithProviders(<Skeleton />);

    expect(container.firstElementChild).toHaveAttribute("data-slot", "skeleton");
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(<Skeleton className="my-skeleton" />);

    expect(container.firstElementChild).toHaveClass("my-skeleton");
  });
});
