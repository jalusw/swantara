import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MultiProgress } from "@/components/multi-progress";
import { renderWithProviders } from "@/lib/tests";

const segments = [
  { id: "ops", label: "Operations", value: 50 },
  { id: "sales", label: "Sales", value: 30 },
  { id: "it", label: "IT", value: 20 },
];

function renderProgress() {
  return renderWithProviders(<MultiProgress ariaLabel="Budget split" segments={segments} />);
}

describe("MultiProgress", () => {
  it("renders the segment labels with computed percentages", () => {
    renderProgress();

    expect(screen.getByText("Operations")).toBeInTheDocument();
    expect(screen.getByText("50%")).toBeInTheDocument();
    expect(screen.getByText("30%")).toBeInTheDocument();
    expect(screen.getByText("20%")).toBeInTheDocument();
  });

  it("sizes each segment proportionally to its share of the total", () => {
    const { container } = renderProgress();
    const bar = container.querySelector('[data-slot="multi-progress"] > div') as HTMLElement;

    const parts = Array.from(bar.children) as HTMLElement[];
    expect(parts).toHaveLength(3);
    expect(parts[0]).toHaveStyle({ width: "50%" });
    expect(parts[1]).toHaveStyle({ width: "30%" });
    expect(parts[2]).toHaveStyle({ width: "20%" });
  });

  it("renders an empty bar when the total is zero", () => {
    const { container } = renderWithProviders(
      <MultiProgress ariaLabel="Empty split" segments={[{ id: "a", label: "A", value: 0 }]} />,
    );
    const bar = container.querySelector('[data-slot="multi-progress"] > div') as HTMLElement;

    expect(bar.children).toHaveLength(0);
    expect(screen.getByText("0%")).toBeInTheDocument();
  });
});
