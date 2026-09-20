import { screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ChartContainer,
  ChartLegendContent,
  ChartStyle,
  ChartTooltipContent,
} from "@/components/chart";
import { renderWithProviders } from "@/lib/tests";

const config = {
  revenue: { label: "Revenue", color: "#3b82f6" },
  cost: { label: "Cost", theme: { light: "#ef4444", dark: "#f87171" } },
};

beforeEach(() => {
  vi.spyOn(Element.prototype, "getBoundingClientRect").mockReturnValue({
    width: 500,
    height: 300,
    top: 0,
    left: 0,
    right: 500,
    bottom: 300,
    x: 0,
    y: 0,
    toJSON: () => {},
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("ChartContainer", () => {
  it("renders the chart slot with themed styles and children", () => {
    renderWithProviders(
      <ChartContainer config={config} className="custom">
        <div>marker</div>
      </ChartContainer>,
    );
    const chart = document.querySelector('[data-slot="chart"]');
    expect(chart).toBeInTheDocument();
    expect(chart).toHaveClass("custom");
    expect(chart?.getAttribute("data-chart")).toMatch(/^chart-/);
    expect(document.querySelector("style")?.innerHTML).toContain("--color-revenue");
    expect(screen.getByText("marker")).toBeInTheDocument();
  });
});

describe("ChartStyle", () => {
  it("renders nothing without colors", () => {
    const { container } = renderWithProviders(<ChartStyle id="chart-1" config={{}} />);
    expect(container.innerHTML).toBe("");
  });
});

describe("ChartTooltipContent", () => {
  it("renders nothing when inactive", () => {
    renderWithProviders(
      <ChartContainer config={config}>
        <ChartTooltipContent active={false} payload={[]} />
      </ChartContainer>,
    );
    expect(document.querySelector('[data-slot="chart"]')).toBeInTheDocument();
    expect(screen.queryByText("Revenue")).not.toBeInTheDocument();
  });

  it("shows the label and value when active", () => {
    renderWithProviders(
      <ChartContainer config={config}>
        <ChartTooltipContent
          active
          payload={[
            {
              name: "revenue",
              value: 1200,
              color: "#3b82f6",
              dataKey: "revenue",
              payload: { revenue: 1200 },
            } as never,
          ]}
        />
      </ChartContainer>,
    );
    expect(screen.getAllByText("Revenue").length).toBeGreaterThan(0);
    expect(screen.getByText("1,200")).toBeInTheDocument();
  });
});

describe("ChartLegendContent", () => {
  it("renders nothing without payload", () => {
    renderWithProviders(
      <ChartContainer config={config}>
        <ChartLegendContent payload={[]} />
      </ChartContainer>,
    );
    expect(document.querySelector('[data-slot="chart"]')).toBeInTheDocument();
    expect(screen.queryByText("Revenue")).not.toBeInTheDocument();
  });

  it("renders configured legend labels", () => {
    renderWithProviders(
      <ChartContainer config={config}>
        <ChartLegendContent
          payload={[{ value: "x", type: "rect", color: "#3b82f6", dataKey: "revenue" }]}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Revenue")).toBeInTheDocument();
  });
});
