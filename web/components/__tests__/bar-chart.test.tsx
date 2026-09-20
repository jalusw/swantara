import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { BarChart } from "@/components/bar-chart";
import { renderWithProviders } from "@/lib/tests";

const data = [
  { label: "Jan", value: 10 },
  { label: "Feb", value: 20 },
  { label: "Mar", value: 15 },
];

describe("BarChart", () => {
  it("renders an accessible chart with labels", () => {
    renderWithProviders(<BarChart data={data} ariaLabel="Sales by month" />);
    expect(screen.getByRole("img", { name: /Sales by month/ })).toBeInTheDocument();
    expect(screen.getByText("Jan")).toBeInTheDocument();
    expect(screen.getByText("Feb")).toBeInTheDocument();
  });

  it("formats values in the accessible description", () => {
    renderWithProviders(
      <BarChart data={data} ariaLabel="Sales" valueFormatter={(value) => `$${value}`} />,
    );
    expect(screen.getByRole("img", { name: /Feb: \$20/ })).toBeInTheDocument();
  });

  it("scales the tallest bar to full height", () => {
    const { container } = renderWithProviders(<BarChart data={data} ariaLabel="Sales" />);
    const bars = container.querySelectorAll('[data-slot="bar-chart"] [title]');
    const heights = [...bars].map((bar) => (bar as HTMLElement).style.height);
    expect(heights).toContain("100%");
  });
});
