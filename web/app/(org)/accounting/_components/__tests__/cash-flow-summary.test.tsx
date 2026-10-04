import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { CashFlowSummary } from "../cash-flow-summary";

describe("CashFlowSummary", () => {
  it("renders cash flow heading with description", () => {
    renderWithProviders(<CashFlowSummary />);

    expect(screen.getByText("Arus kas")).toBeInTheDocument();
    expect(screen.getByText("Pemasukan dan pengeluaran kas bulan ini.")).toBeInTheDocument();
  });

  it("renders inflow, outflow and net rows", () => {
    renderWithProviders(<CashFlowSummary />);

    expect(screen.getByText("Pemasukan")).toBeInTheDocument();
    expect(screen.getByText("Pengeluaran")).toBeInTheDocument();
    expect(screen.getByText("Arus kas bersih")).toBeInTheDocument();
  });
});
