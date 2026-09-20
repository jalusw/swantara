import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { CashFlowSummary } from "../cash-flow-summary";

describe("CashFlowSummary", () => {
  it("renders cash flow heading with description", () => {
    renderWithProviders(<CashFlowSummary />);

    expect(screen.getByText("Cash flow")).toBeInTheDocument();
    expect(screen.getByText("Money in and out this month.")).toBeInTheDocument();
  });

  it("renders inflow, outflow and net rows", () => {
    renderWithProviders(<CashFlowSummary />);

    expect(screen.getByText("Inflows")).toBeInTheDocument();
    expect(screen.getByText("Outflows")).toBeInTheDocument();
    expect(screen.getByText("Net")).toBeInTheDocument();
  });
});
