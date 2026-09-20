import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/dropdown-menu";
import { renderWithProviders } from "@/lib/tests";

function renderOpenMenu(children: React.ReactNode) {
  return renderWithProviders(
    <DropdownMenu open>
      <DropdownMenuTrigger>Menu</DropdownMenuTrigger>
      <DropdownMenuContent>{children}</DropdownMenuContent>
    </DropdownMenu>,
  );
}

describe("DropdownMenu", () => {
  it("renders a trigger with the menu slot", () => {
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

  it("renders label, item, separator, and shortcut when open", () => {
    renderOpenMenu(
      <>
        <DropdownMenuGroup>
          <DropdownMenuLabel>Actions</DropdownMenuLabel>
          <DropdownMenuItem>
            Edit <DropdownMenuShortcut>⌘E</DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem variant="destructive">Delete</DropdownMenuItem>
      </>,
    );
    expect(screen.getByText("Actions")).toHaveAttribute("data-slot", "dropdown-menu-label");
    expect(screen.getByText("Delete")).toHaveAttribute("data-variant", "destructive");
    expect(screen.getByText("⌘E")).toHaveAttribute("data-slot", "dropdown-menu-shortcut");
    expect(document.querySelector("[data-slot='dropdown-menu-separator']")).toBeInTheDocument();
  });

  it("activates an item on click", async () => {
    const onClick = vi.fn();
    renderOpenMenu(<DropdownMenuItem onClick={onClick}>Edit</DropdownMenuItem>);
    await userEvent.setup().click(screen.getByText("Edit"));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("renders a group wrapper", () => {
    renderOpenMenu(
      <DropdownMenuGroup>
        <DropdownMenuItem>Grouped</DropdownMenuItem>
      </DropdownMenuGroup>,
    );
    expect(document.querySelector("[data-slot='dropdown-menu-group']")).toBeInTheDocument();
    expect(screen.getByText("Grouped")).toBeInTheDocument();
  });

  it("renders a submenu trigger with its content", () => {
    renderOpenMenu(
      <DropdownMenuSub open>
        <DropdownMenuSubTrigger>More</DropdownMenuSubTrigger>
        <DropdownMenuSubContent>
          <DropdownMenuItem>Nested</DropdownMenuItem>
        </DropdownMenuSubContent>
      </DropdownMenuSub>,
    );
    expect(screen.getByText("More")).toHaveAttribute("data-slot", "dropdown-menu-sub-trigger");
    expect(screen.getByText("Nested")).toBeInTheDocument();
  });

  it("renders a checked checkbox item", () => {
    renderOpenMenu(<DropdownMenuCheckboxItem checked>Show grid</DropdownMenuCheckboxItem>);
    expect(screen.getByText("Show grid")).toBeInTheDocument();
    expect(
      document.querySelector("[data-slot='dropdown-menu-checkbox-item-indicator']"),
    ).toBeInTheDocument();
  });

  it("renders a radio group with items", () => {
    renderOpenMenu(
      <DropdownMenuRadioGroup value="a">
        <DropdownMenuRadioItem value="a">Option A</DropdownMenuRadioItem>
        <DropdownMenuRadioItem value="b">Option B</DropdownMenuRadioItem>
      </DropdownMenuRadioGroup>,
    );
    expect(screen.getByText("Option A")).toBeInTheDocument();
    expect(screen.getByText("Option B")).toBeInTheDocument();
    expect(document.querySelector("[data-slot='dropdown-menu-radio-group']")).toBeInTheDocument();
  });

  it("renders a portal wrapper", () => {
    renderWithProviders(
      <DropdownMenu open>
        <DropdownMenuTrigger>Menu</DropdownMenuTrigger>
        <DropdownMenuPortal>
          <DropdownMenuContent>
            <DropdownMenuItem>Portaled</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenuPortal>
      </DropdownMenu>,
    );
    expect(screen.getByText("Portaled")).toBeInTheDocument();
  });

  it("supports an asChild trigger", () => {
    renderWithProviders(
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button type="button" className="custom-trigger">
            Menu
          </button>
        </DropdownMenuTrigger>
      </DropdownMenu>,
    );
    expect(screen.getByRole("button", { name: "Menu" })).toHaveClass("custom-trigger");
  });
});
