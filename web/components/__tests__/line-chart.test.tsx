import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { LineChart } from "@/components/line-chart";
import { renderWithProviders } from "@/lib/tests";

const data = [
  { label: "Jan", value: 10 },
  { label: "Feb", value: 24 },
  { label: "Mar", value: 18 },
];

describe("LineChart", () => {
  it("renders an accessible chart with a title and label axis", () => {
    renderWithProviders(<LineChart data={data} ariaLabel="Revenue trend" />);
    expect(screen.getByLabelText("Revenue trend")).toBeInTheDocument();
    expect(screen.getByTitle("Revenue trend")).toBeInTheDocument();
    expect(screen.getByText("Feb")).toBeInTheDocument();
  });

  it("renders point labels with the value formatter", () => {
    renderWithProviders(
      <LineChart
        data={data}
        ariaLabel="Revenue trend"
        variant="line"
        valueFormatter={(value) => `$${value}`}
      />,
    );
    expect(screen.getByTitle("Feb: $24")).toBeInTheDocument();
  });
});
