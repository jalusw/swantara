import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Popover,
  PopoverContent,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from "@/components/popover";
import { renderWithProviders } from "@/lib/tests";

describe("Popover", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <Popover>
        <PopoverTrigger>Toggle</PopoverTrigger>
      </Popover>,
    );

    expect(screen.getByRole("button", { name: "Toggle" })).toHaveAttribute(
      "data-slot",
      "popover-trigger",
    );
  });

  it("renders content with header and title when open", () => {
    renderWithProviders(
      <Popover open>
        <PopoverTrigger>Toggle</PopoverTrigger>
        <PopoverContent>
          <PopoverHeader>
            <PopoverTitle>Heading</PopoverTitle>
          </PopoverHeader>
          <div>Body</div>
        </PopoverContent>
      </Popover>,
    );

    expect(screen.getByText("Heading")).toHaveAttribute("data-slot", "popover-title");
    expect(screen.getByText("Body")).toBeInTheDocument();
  });

  it("renders header with data-slot", () => {
    renderWithProviders(<PopoverHeader data-slot="popover-header">Info</PopoverHeader>);

    expect(screen.getByText("Info")).toHaveAttribute("data-slot", "popover-header");
  });
});
