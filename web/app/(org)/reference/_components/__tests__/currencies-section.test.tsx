import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { CurrenciesSection } from "../currencies-section";

beforeEach(() => {});

describe("CurrenciesSection", () => {
  it("renders the seeded currencies", async () => {
    renderWithProviders(<CurrenciesSection orgId="1" />);

    expect(await screen.findByText("US Dollar")).toBeInTheDocument();
    expect(screen.getByText("Indonesian Rupiah")).toBeInTheDocument();
  });

  it("is read-only and offers no add action", () => {
    renderWithProviders(<CurrenciesSection orgId="1" />);

    expect(screen.queryByRole("button", { name: /add/i })).not.toBeInTheDocument();
  });
});
