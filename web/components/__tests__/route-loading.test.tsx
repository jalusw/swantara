import { describe, expect, it } from "vitest";
import { RouteLoading } from "@/components/route-loading";
import { renderWithProviders } from "@/lib/tests";

describe("RouteLoading", () => {
  it("renders page variant with spinner icon", () => {
    const { container } = renderWithProviders(<RouteLoading />);

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
  });

  it("renders table variant with skeleton rows", () => {
    const { container } = renderWithProviders(<RouteLoading variant="table" rows={4} />);

    const rows = container.querySelectorAll(".animate-pulse");
    expect(rows).toHaveLength(4);
  });

  it("defaults to 3 rows for table variant", () => {
    const { container } = renderWithProviders(<RouteLoading variant="table" />);

    const rows = container.querySelectorAll(".animate-pulse");
    expect(rows).toHaveLength(3);
  });
});
