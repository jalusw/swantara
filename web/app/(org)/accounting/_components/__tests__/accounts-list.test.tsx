import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { AccountsList } from "../accounts-list";

const accounts = [
  { name: "Current assets", balance: 58200, movement: "up", percent: "+4.2%" },
  { name: "Accounts payable", balance: 12800, movement: "down", percent: "-1.1%" },
  { name: "Accounts receivable", balance: 41300, movement: "up", percent: "+6.5%" },
] as const;

describe("AccountsList", () => {
  it("renders account names with balances", async () => {
    renderWithProviders(<AccountsList accounts={[...accounts]} />);

    expect(await screen.findByText("Current assets")).toBeInTheDocument();
    expect(screen.getByText("Accounts payable")).toBeInTheDocument();
    expect(screen.getByText("IDR 58,200.00")).toBeInTheDocument();
    expect(screen.getByText("IDR 12,800.00")).toBeInTheDocument();
  });

  it("renders movement percentages", () => {
    renderWithProviders(<AccountsList accounts={[...accounts]} />);

    expect(screen.getByText("+4.2%")).toBeInTheDocument();
    expect(screen.getByText("-1.1%")).toBeInTheDocument();
    expect(screen.getByText("+6.5%")).toBeInTheDocument();
  });
});
