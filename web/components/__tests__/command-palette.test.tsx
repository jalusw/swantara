import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CommandPaletteProvider, useCommandPalette } from "@/components/command-palette";
import { renderWithProviders } from "@/lib/tests";

function PaletteStatus() {
  const { open } = useCommandPalette();
  return <span>{open ? "open" : "closed"}</span>;
}

describe("CommandPalette", () => {
  it("renders the provider wrapper with data-slot", () => {
    renderWithProviders(
      <CommandPaletteProvider>
        <span>child</span>
      </CommandPaletteProvider>,
    );

    expect(screen.getByText("child").closest("[data-slot='command-palette']")).toBeInTheDocument();
  });

  it("provides closed state by default", () => {
    renderWithProviders(
      <CommandPaletteProvider>
        <PaletteStatus />
      </CommandPaletteProvider>,
    );

    expect(screen.getByText("closed")).toBeInTheDocument();
  });

  it("throws when useCommandPalette is used outside provider", () => {
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => renderWithProviders(<PaletteStatus />)).toThrow(
      "useCommandPalette must be used within a CommandPaletteProvider",
    );

    spy.mockRestore();
  });
});
