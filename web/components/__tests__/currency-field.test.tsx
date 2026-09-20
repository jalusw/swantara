import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  CurrencyField,
  CurrencyView,
  formatCurrency,
  parseCurrency,
} from "@/components/currency-field";
import { renderWithProviders } from "@/lib/tests";

describe("formatCurrency", () => {
  it("formats a number for the given locale and currency", () => {
    expect(formatCurrency(1250, { locale: "en-US", currency: "USD" })).toBe("$1,250.00");
  });
});

describe("parseCurrency", () => {
  it("parses symbols and grouping into a number", () => {
    expect(parseCurrency("$1,250.00")).toBe(1250);
    expect(parseCurrency("1,500")).toBe(1500);
    expect(parseCurrency("")).toBeNull();
  });
});

describe("CurrencyView", () => {
  it("shows a formatted value", () => {
    render(<CurrencyView value={1250} currency="USD" />);
    expect(screen.getByText("$1,250.00")).toBeInTheDocument();
  });

  it("shows a placeholder when null", () => {
    render(<CurrencyView value={null} />);
    expect(screen.getByText("—")).toBeInTheDocument();
  });
});

describe("CurrencyField", () => {
  it("reports a parsed number on input", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    renderWithProviders(<CurrencyField value={null} onValueChange={onValueChange} />);
    const input = screen.getByRole("textbox");
    await user.type(input, "1,500");
    expect(onValueChange).toHaveBeenLastCalledWith(1500);
  });
});
