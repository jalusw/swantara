import { screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartStyle,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/chart";
import { renderWithProviders } from "@/lib/tests";

const BASE_CONFIG = {
  revenue: { label: "Revenue", color: "#3b82f6" },
  cost: { label: "Cost", theme: { light: "#ef4444", dark: "#f87171" } },
};

const ICON_CONFIG = {
  sales: {
    label: "Sales",
    color: "#22c55e",
    icon: () => <span data-testid="tooltip-icon">icon</span>,
  },
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

describe("ChartContainer branches", () => {
  it("uses an explicit id for the chart slot", () => {
    renderWithProviders(
      <ChartContainer id="sales" config={BASE_CONFIG}>
        <div>body</div>
      </ChartContainer>,
    );
    expect(document.querySelector('[data-chart="chart-sales"]')).toBeInTheDocument();
  });

  it("accepts a custom initial dimension", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG} initialDimension={{ width: 640, height: 400 }}>
        <div>body</div>
      </ChartContainer>,
    );
    expect(document.querySelector('[data-slot="chart"]')).toBeInTheDocument();
  });
});

describe("ChartStyle branches", () => {
  it("emits theme variables for themed entries", () => {
    renderWithProviders(<ChartStyle id="chart-themed" config={BASE_CONFIG} />);
    const html = document.querySelector("style")?.innerHTML ?? "";
    expect(html).toContain("--color-revenue");
    expect(html).toContain("--color-cost");
    expect(html).toContain(".dark [data-chart=chart-themed]");
  });

  it("skips entries without color or theme", () => {
    const { container } = renderWithProviders(
      <ChartStyle id="chart-empty" config={{ plain: { label: "Plain" } }} />,
    );
    expect(container.innerHTML).toBe("");
  });

  it("emits only the entries that carry a color", () => {
    renderWithProviders(
      <ChartStyle
        id="chart-mixed"
        config={{ plain: { label: "Plain" }, revenue: { label: "Revenue", color: "#000" } }}
      />,
    );
    const html = document.querySelector("style")?.innerHTML ?? "";
    expect(html).toContain("--color-revenue");
    expect(html).not.toContain("--color-plain");
  });
});

describe("ChartTooltipContent branches", () => {
  const payload = (overrides = {}) => [
    {
      name: "revenue",
      value: 1200,
      color: "#3b82f6",
      dataKey: "revenue",
      payload: { revenue: 1200 },
      ...overrides,
    } as never,
  ];

  it("returns null when payload is missing", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={undefined} />
      </ChartContainer>,
    );
    expect(screen.queryByText("Revenue")).not.toBeInTheDocument();
  });

  it("hides the label when hideLabel is set", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={payload()} hideLabel />
      </ChartContainer>,
    );
    expect(screen.getByText("1.200")).toBeInTheDocument();
  });

  it("formats the label through labelFormatter", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent
          active
          payload={payload()}
          label="revenue"
          labelFormatter={(value) => `Total ${value}`}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Total Revenue")).toBeInTheDocument();
  });

  it("resolves the label from a labelKey", () => {
    renderWithProviders(
      <ChartContainer config={{ day: { label: "Day" }, revenue: { label: "Revenue" } }}>
        <ChartTooltipContent
          active
          payload={[{ name: "x", value: 5, dataKey: "revenue", payload: {} } as never]}
          labelKey="day"
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Day")).toBeInTheDocument();
  });

  it("renders nothing for the label when no value resolves", () => {
    renderWithProviders(
      <ChartContainer config={{}}>
        <ChartTooltipContent active payload={payload()} label="missing" />
      </ChartContainer>,
    );
    expect(screen.getByText("1.200")).toBeInTheDocument();
  });

  it("uses a custom formatter for the item row", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent
          active
          payload={payload()}
          formatter={(value) => <span>custom-{String(value)}</span>}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("custom-1200")).toBeInTheDocument();
  });

  it("skips payload entries typed none", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={payload({ type: "none" })} />
      </ChartContainer>,
    );
    expect(screen.queryByText("1.200")).not.toBeInTheDocument();
  });

  it("renders the configured icon instead of the indicator", () => {
    renderWithProviders(
      <ChartContainer config={ICON_CONFIG}>
        <ChartTooltipContent
          active
          payload={[
            {
              name: "sales",
              value: 42,
              color: "#22c55e",
              dataKey: "sales",
              payload: { sales: 42 },
            } as never,
          ]}
        />
      </ChartContainer>,
    );
    expect(screen.getByTestId("tooltip-icon")).toBeInTheDocument();
  });

  it("renders a line indicator with nested label", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={payload()} indicator="line" />
      </ChartContainer>,
    );
    expect(screen.getAllByText("Revenue").length).toBeGreaterThan(0);
  });

  it("renders a dashed indicator", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={payload()} indicator="dashed" />
      </ChartContainer>,
    );
    expect(screen.getAllByText("Revenue").length).toBeGreaterThan(0);
  });

  it("hides the indicator when requested", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={payload()} hideIndicator />
      </ChartContainer>,
    );
    expect(screen.getByText("1.200")).toBeInTheDocument();
  });

  it("renders string values without locale formatting", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent
          active
          payload={payload({ name: "note", value: "n/a", dataKey: "note", payload: {} })}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("n/a")).toBeInTheDocument();
  });

  it("omits the value span when value is null", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartTooltipContent active payload={payload({ value: null })} />
      </ChartContainer>,
    );
    expect(document.querySelector('[data-slot="chart"]')).toBeInTheDocument();
  });

  it("resolves config through a nested payload key", () => {
    renderWithProviders(
      <ChartContainer config={{ apples: { label: "Apples" } }}>
        <ChartTooltipContent
          active
          payload={[{ name: "x", value: 3, payload: { fruit: "apples" } } as never]}
          nameKey="fruit"
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Apples")).toBeInTheDocument();
  });

  it("resolves config through a direct payload key", () => {
    renderWithProviders(
      <ChartContainer config={{ apples: { label: "Apples" } }}>
        <ChartTooltipContent
          active
          payload={[{ name: "x", value: 3, fruit: "apples", payload: {} } as never]}
          nameKey="fruit"
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Apples")).toBeInTheDocument();
  });

  it("falls back to item name when config is missing", () => {
    renderWithProviders(
      <ChartContainer config={{}}>
        <ChartTooltipContent
          active
          payload={[{ name: "mystery", value: 7, payload: {} } as never]}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("mystery")).toBeInTheDocument();
  });

  it("throws when tooltip content renders outside a container", () => {
    expect(() => renderWithProviders(<ChartTooltipContent active payload={payload()} />)).toThrow(
      "useChart must be used within a <ChartContainer />",
    );
  });

  it("re-exports the recharts tooltip primitives", () => {
    expect(ChartTooltip).toBeDefined();
    expect(ChartLegend).toBeDefined();
  });
});

describe("ChartLegendContent branches", () => {
  it("renders nothing when payload is missing", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartLegendContent payload={undefined} />
      </ChartContainer>,
    );
    expect(screen.queryByText("Revenue")).not.toBeInTheDocument();
  });

  it("applies top alignment spacing", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartLegendContent
          verticalAlign="top"
          payload={[{ value: "x", type: "rect", color: "#3b82f6", dataKey: "revenue" }]}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Revenue")).toBeInTheDocument();
  });

  it("hides the icon when requested", () => {
    renderWithProviders(
      <ChartContainer config={ICON_CONFIG}>
        <ChartLegendContent
          hideIcon
          payload={[{ value: "x", type: "rect", color: "#22c55e", dataKey: "sales" }]}
        />
      </ChartContainer>,
    );
    expect(screen.queryByTestId("tooltip-icon")).not.toBeInTheDocument();
    expect(screen.getByText("Sales")).toBeInTheDocument();
  });

  it("renders the configured legend icon", () => {
    renderWithProviders(
      <ChartContainer config={ICON_CONFIG}>
        <ChartLegendContent
          payload={[{ value: "x", type: "rect", color: "#22c55e", dataKey: "sales" }]}
        />
      </ChartContainer>,
    );
    expect(screen.getByTestId("tooltip-icon")).toBeInTheDocument();
  });

  it("skips legend entries typed none", () => {
    renderWithProviders(
      <ChartContainer config={BASE_CONFIG}>
        <ChartLegendContent
          payload={[{ value: "x", type: "none", color: "#3b82f6", dataKey: "revenue" }]}
        />
      </ChartContainer>,
    );
    expect(screen.queryByText("Revenue")).not.toBeInTheDocument();
  });

  it("resolves legend config through nameKey", () => {
    renderWithProviders(
      <ChartContainer config={{ apples: { label: "Apples" } }}>
        <ChartLegendContent
          nameKey="fruit"
          payload={[{ value: "x", type: "rect", color: "#000", payload: { fruit: "apples" } }]}
        />
      </ChartContainer>,
    );
    expect(screen.getByText("Apples")).toBeInTheDocument();
  });
});
