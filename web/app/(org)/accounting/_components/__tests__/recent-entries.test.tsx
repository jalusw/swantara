import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { RecentEntries } from "../recent-entries";

const entries = [
  { date: "2026-08-05", description: "Payment received", debit: null, credit: 2400, tag: "income" },
  { date: "2026-08-04", description: "Payroll run", debit: 1940, credit: null, tag: "expense" },
  { date: "2026-08-03", description: "Supplier payment", debit: 880, credit: null, tag: "expense" },
] as const;

describe("RecentEntries", () => {
  it("renders entry descriptions with dates", async () => {
    renderWithProviders(<RecentEntries entries={[...entries]} />);

    expect(await screen.findByText("Payment received")).toBeInTheDocument();
    expect(screen.getByText("Payroll run")).toBeInTheDocument();
    expect(screen.getByText("2026-08-05")).toBeInTheDocument();
  });

  it("renders income and expense tags", () => {
    renderWithProviders(<RecentEntries entries={[...entries]} />);

    expect(screen.getByText("IN")).toBeInTheDocument();
    expect(screen.getAllByText("EX").length).toBeGreaterThan(0);
  });

  it("renders debit and credit amounts", () => {
    renderWithProviders(<RecentEntries entries={[...entries]} />);

    expect(screen.getByText("Rp 2.400,00")).toBeInTheDocument();
    expect(screen.getByText("Rp 1.940,00")).toBeInTheDocument();
  });
});
