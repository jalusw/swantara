import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MetricGrid } from "@/components/metric-grid";
import { renderWithProviders } from "@/lib/tests";

describe("MetricGrid", () => {
  it("renders its children in a responsive grid", () => {
    const { container } = renderWithProviders(
      <MetricGrid>
        <span>One</span>
        <span>Two</span>
      </MetricGrid>,
    );
    expect(screen.getByText("One")).toBeInTheDocument();
    expect(screen.getByText("Two")).toBeInTheDocument();
    expect(container.querySelector('[data-slot="metric-grid"]')).toHaveClass("grid");
  });
});
