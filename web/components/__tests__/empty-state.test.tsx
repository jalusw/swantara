import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { EmptyState } from "@/components/empty-state";
import { renderWithProviders } from "@/lib/tests";

describe("EmptyState", () => {
  it("renders title, description, and actions", () => {
    renderWithProviders(
      <EmptyState title="No customers yet" description="Create your first customer.">
        <button type="button">Add customer</button>
      </EmptyState>,
    );

    expect(screen.getByRole("heading", { name: "No customers yet" })).toBeInTheDocument();
    expect(screen.getByText("Create your first customer.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add customer" })).toBeInTheDocument();
  });

  it("omits description when not provided", () => {
    renderWithProviders(<EmptyState title="Nothing here" />);
    expect(screen.getByRole("heading", { name: "Nothing here" })).toBeInTheDocument();
  });
});
