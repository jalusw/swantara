import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { BackLink } from "@/components/back-link";
import { renderWithProviders } from "@/lib/tests";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn() }),
}));

describe("BackLink", () => {
  it("renders a button with the given text", () => {
    renderWithProviders(<BackLink href="/home">Go back</BackLink>);

    expect(screen.getByRole("button", { name: /go back/i })).toBeInTheDocument();
  });

  it("has the data-slot attribute", () => {
    renderWithProviders(<BackLink href="/dashboard">Back</BackLink>);

    expect(screen.getByRole("button", { name: "Back" })).toHaveAttribute("data-slot", "back-link");
  });
});
