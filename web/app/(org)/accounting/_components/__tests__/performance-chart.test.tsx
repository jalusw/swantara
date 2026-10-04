import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PerformanceChart } from "../performance-chart";

describe("PerformanceChart", () => {
  it("renders performance heading with default net profit chart", () => {
    renderWithProviders(<PerformanceChart />);

    expect(screen.getByText("Kinerja")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: /Laba bersih/ })).toBeInTheDocument();
  });

  it("switches to revenue chart on tab select", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PerformanceChart />);

    await user.click(screen.getByRole("tab", { name: "Pendapatan" }));

    expect(await screen.findByRole("img", { name: /^Pendapatan\./ })).toBeInTheDocument();
  });

  it("switches to expenses chart on tab select", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PerformanceChart />);

    await user.click(screen.getByRole("tab", { name: "Biaya" }));

    expect(await screen.findByRole("img", { name: /^Biaya\./ })).toBeInTheDocument();
  });
});
