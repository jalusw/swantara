import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SkipToMain } from "@/components/skip-to-main";
import { renderWithProviders } from "@/lib/tests";

describe("SkipToMain", () => {
  it("renders a skip link with sensible defaults", () => {
    renderWithProviders(<SkipToMain />);

    const link = screen.getByRole("link", { name: "Skip to main content" });
    expect(link).toHaveAttribute("href", "#main");
  });

  it("honours custom label and href", () => {
    renderWithProviders(<SkipToMain label="Skip nav" href="#content" />);

    const link = screen.getByRole("link", { name: "Skip nav" });
    expect(link).toHaveAttribute("href", "#content");
  });
});
