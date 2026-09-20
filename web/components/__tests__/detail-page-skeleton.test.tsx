import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import { renderWithProviders } from "@/lib/tests";

describe("DetailPageSkeleton", () => {
  it("renders placeholder skeletons", () => {
    const { container } = renderWithProviders(<DetailPageSkeleton />);

    expect(container.querySelectorAll("[data-slot='skeleton']").length).toBeGreaterThan(0);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});
