import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/command";
import { renderWithProviders } from "@/lib/tests";

describe("Command", () => {
  it("renders with data-slot", () => {
    renderWithProviders(
      <Command>
        <CommandInput placeholder="Search commands..." />
      </Command>,
    );

    expect(
      screen.getByPlaceholderText("Search commands...").closest("[data-slot='command']"),
    ).toBeInTheDocument();
  });

  it("renders items inside a group", () => {
    renderWithProviders(
      <Command>
        <CommandList>
          <CommandGroup heading="Actions">
            <CommandItem>Copy</CommandItem>
            <CommandItem>Paste</CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>,
    );

    expect(screen.getByText("Copy")).toHaveAttribute("data-slot", "command-item");
    expect(screen.getByText("Paste")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    renderWithProviders(
      <Command>
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
        </CommandList>
      </Command>,
    );

    expect(screen.getByText("No results found.")).toHaveAttribute("data-slot", "command-empty");
  });
});
