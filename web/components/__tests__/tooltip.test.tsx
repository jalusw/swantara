import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/tooltip";
import { renderWithProviders } from "@/lib/tests";

describe("Tooltip", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <Tooltip>
        <TooltipTrigger>Hover me</TooltipTrigger>
      </Tooltip>,
    );

    expect(screen.getByRole("button", { name: "Hover me" })).toHaveAttribute(
      "data-slot",
      "tooltip-trigger",
    );
  });

  it("renders content with data-slot when open", () => {
    renderWithProviders(
      <Tooltip open>
        <TooltipTrigger>Hover me</TooltipTrigger>
        <TooltipContent>Tooltip text</TooltipContent>
      </Tooltip>,
    );

    expect(screen.getByText("Tooltip text")).toHaveAttribute("data-slot", "tooltip-content");
  });

  it("passes through a custom className on content", () => {
    renderWithProviders(
      <Tooltip open>
        <TooltipTrigger>Hover me</TooltipTrigger>
        <TooltipContent className="my-tooltip">Text</TooltipContent>
      </Tooltip>,
    );

    expect(screen.getByText("Text")).toHaveClass("my-tooltip");
  });
});
