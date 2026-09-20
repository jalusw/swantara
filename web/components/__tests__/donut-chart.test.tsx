import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DonutChart } from "@/components/donut-chart";
import { renderWithProviders } from "@/lib/tests";

const data = [
  { label: "Software", value: 42000 },
  { label: "Hardware", value: 28000 },
  { label: "Services", value: 15000 },
];

describe("DonutChart", () => {
  it("renders an accessible chart with a legend", () => {
    renderWithProviders(<DonutChart data={data} ariaLabel="Revenue mix" />);

    expect(screen.getByLabelText(/Revenue mix/)).toBeInTheDocument();
    expect(screen.getByText("Software")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getByText("85000")).toBeInTheDocument();
  });

  it("formats values and shows percentages", () => {
    renderWithProviders(
      <DonutChart
        data={data}
        ariaLabel="Revenue mix"
        valueFormatter={(value) => `$${value.toLocaleString()}`}
      />,
    );

    expect(screen.getByText("$85,000")).toBeInTheDocument();
    expect(screen.getAllByText(/\(49%\)/).length).toBeGreaterThan(0);
    expect(screen.getByText("$42,000")).toBeInTheDocument();
  });

  it("shows a no-data state for empty input", () => {
    renderWithProviders(<DonutChart data={[]} ariaLabel="Revenue mix" />);

    expect(screen.getByText("No data")).toBeInTheDocument();
    expect(screen.getByLabelText(/No data/)).toBeInTheDocument();
  });
});
