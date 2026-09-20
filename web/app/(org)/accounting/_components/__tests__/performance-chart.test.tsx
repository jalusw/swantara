import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PerformanceChart } from "../performance-chart";

describe("PerformanceChart", () => {
  it("renders performance heading with default net profit chart", () => {
    renderWithProviders(<PerformanceChart />);

    expect(screen.getByText("Performance")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: /Net profit/ })).toBeInTheDocument();
  });

  it("switches to revenue chart on tab select", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PerformanceChart />);

    await user.click(screen.getByRole("tab", { name: "Revenue" }));

    expect(await screen.findByRole("img", { name: /^Revenue\./ })).toBeInTheDocument();
  });

  it("switches to expenses chart on tab select", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PerformanceChart />);

    await user.click(screen.getByRole("tab", { name: "Expenses" }));

    expect(await screen.findByRole("img", { name: /^Expenses\./ })).toBeInTheDocument();
  });
});
