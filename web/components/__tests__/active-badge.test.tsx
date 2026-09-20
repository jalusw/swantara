import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ActiveBadge } from "@/components/active-badge";
import { renderWithProviders } from "@/lib/tests";

describe("ActiveBadge", () => {
  it("renders children with the default variant when active", () => {
    renderWithProviders(<ActiveBadge active>Active</ActiveBadge>);

    const badge = screen.getByText("Active");
    expect(badge).toBeInTheDocument();
    expect(badge.closest("[data-slot='badge']")).not.toBeNull();
  });

  it("renders children with the outline variant when inactive", () => {
    renderWithProviders(<ActiveBadge active={false}>Inactive</ActiveBadge>);

    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });
});
