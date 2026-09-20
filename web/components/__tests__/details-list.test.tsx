import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { DetailsList } from "@/components/details-list";
import { renderWithProviders } from "@/lib/tests";

const ITEMS = [
  { id: "email", label: "Email", value: "billing@acme.com" },
  { id: "terms", label: "Payment terms", value: "Net 30" },
];

describe("DetailsList", () => {
  it("renders label/value rows", () => {
    renderWithProviders(<DetailsList items={ITEMS} />);

    expect(screen.getByText("Email")).toBeInTheDocument();
    expect(screen.getByText("billing@acme.com")).toBeInTheDocument();
    expect(screen.getByText("Payment terms")).toBeInTheDocument();
    expect(screen.getByText("Net 30")).toBeInTheDocument();
  });

  it("renders the grid layout", () => {
    renderWithProviders(<DetailsList items={ITEMS} layout="grid" />);

    expect(screen.getByText("Email")).toBeInTheDocument();
    expect(screen.getByText("billing@acme.com")).toBeInTheDocument();
  });

  it("supports three columns in grid layout", () => {
    const { container } = renderWithProviders(
      <DetailsList items={ITEMS} layout="grid" columns={3} />,
    );

    expect(container.querySelector("[data-slot='details-list']")).toHaveClass("sm:grid-cols-3");
  });
});
