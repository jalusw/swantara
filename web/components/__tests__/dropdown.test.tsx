import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/dropdown-menu";
import { renderWithProviders } from "@/lib/tests";

describe("DropdownMenu", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <DropdownMenu>
        <DropdownMenuTrigger>Menu</DropdownMenuTrigger>
      </DropdownMenu>,
    );

    expect(screen.getByRole("button", { name: "Menu" })).toHaveAttribute(
      "data-slot",
      "dropdown-menu-trigger",
    );
  });

  it("renders items when open", () => {
    renderWithProviders(
      <DropdownMenu open>
        <DropdownMenuTrigger>Menu</DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem>Edit</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>,
    );

    expect(screen.getByText("Edit")).toHaveAttribute("data-slot", "dropdown-menu-item");
  });

  it("passes through a custom className on trigger", () => {
    renderWithProviders(
      <DropdownMenu>
        <DropdownMenuTrigger className="my-trigger">Menu</DropdownMenuTrigger>
      </DropdownMenu>,
    );

    expect(screen.getByRole("button", { name: "Menu" })).toHaveClass("my-trigger");
  });
});
