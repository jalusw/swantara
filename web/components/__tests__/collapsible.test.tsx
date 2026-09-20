import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/collapsible";
import { renderWithProviders } from "@/lib/tests";

describe("Collapsible", () => {
  it("renders the collapsible container with data-slot", () => {
    const { container } = renderWithProviders(
      <Collapsible>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Hidden content</CollapsibleContent>
      </Collapsible>,
    );

    expect(container.querySelector('[data-slot="collapsible"]')).toBeInTheDocument();
  });

  it("renders the trigger button", () => {
    renderWithProviders(
      <Collapsible>
        <CollapsibleTrigger>Show more</CollapsibleTrigger>
        <CollapsibleContent>Details</CollapsibleContent>
      </Collapsible>,
    );

    expect(screen.getByRole("button", { name: "Show more" })).toHaveAttribute(
      "data-slot",
      "collapsible-trigger",
    );
  });

  it("renders the content panel", () => {
    const { container } = renderWithProviders(
      <Collapsible open>
        <CollapsibleTrigger>Toggle</CollapsibleTrigger>
        <CollapsibleContent>Panel content</CollapsibleContent>
      </Collapsible>,
    );

    expect(container.querySelector('[data-slot="collapsible-content"]')).toHaveTextContent(
      "Panel content",
    );
  });
});
