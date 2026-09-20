import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StateBadge } from "@/components/state-badge";
import { renderWithProviders } from "@/lib/tests";

const statuses = {
  draft: { label: "Draft", tone: "neutral" },
  paid: { label: "Paid", tone: "success" },
} as const;

describe("StateBadge", () => {
  it("renders the mapped label for a known status", () => {
    renderWithProviders(<StateBadge value="paid" statuses={statuses} />);
    expect(screen.getByText("Paid")).toBeInTheDocument();
  });

  it("renders nothing for an unknown status", () => {
    const { container } = renderWithProviders(<StateBadge value="nope" statuses={statuses} />);
    expect(container).toBeEmptyDOMElement();
  });
});
