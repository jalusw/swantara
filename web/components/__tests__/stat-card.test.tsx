import { screen } from "@testing-library/react";
import { Users } from "lucide-react";
import { describe, expect, it } from "vitest";
import { StatCard } from "@/components/stat-card";
import { renderWithProviders } from "@/lib/tests";

describe("StatCard", () => {
  it("renders the label, value and trend", () => {
    renderWithProviders(
      <StatCard label="Customers" value="1,284" icon={Users} trend="+12% from last month" />,
    );

    expect(screen.getByText("Customers")).toBeInTheDocument();
    expect(screen.getByText("1,284")).toBeInTheDocument();
    expect(screen.getByText("+12% from last month")).toBeInTheDocument();
  });

  it("renders a destructive trend icon when the direction is down", () => {
    const { container } = renderWithProviders(
      <StatCard
        label="Employees"
        value="48"
        icon={Users}
        trend="-2% from last month"
        trendDirection="down"
      />,
    );

    expect(container.querySelector("svg.text-destructive")).toBeInTheDocument();
    expect(container.querySelector("svg.text-success")).toBeNull();
  });

  it("omits the trend block when no trend is provided", () => {
    renderWithProviders(<StatCard label="Revenue" value="$84k" icon={Users} />);

    expect(screen.getByText("Revenue")).toBeInTheDocument();
    expect(screen.queryByText("from last month")).not.toBeInTheDocument();
  });
});
