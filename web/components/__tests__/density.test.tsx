import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { act } from "react";
import { describe, expect, it } from "vitest";

import { renderHookWithProviders, renderWithProviders } from "@/lib/tests";
import { COMFORTABLE, COMPACT, NORMAL, useDensityStore } from "@/stores/density.store";
import { DensityToggle, useDataDensity } from "../density";

describe("DensityToggle", () => {
  it("renders the three density options", () => {
    renderWithProviders(<DensityToggle />);

    expect(screen.getByRole("button", { name: "Compact" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Normal" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Comfortable" })).toBeInTheDocument();
  });

  it("starts at normal", () => {
    renderWithProviders(<DensityToggle />);

    expect(useDensityStore.getState().density).toBe(NORMAL);
  });

  it("updates the store and persists when an option is selected", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DensityToggle />);

    await user.click(screen.getByRole("button", { name: "Compact" }));

    expect(useDensityStore.getState().density).toBe(COMPACT);
    expect(localStorage.getItem("data-density")).toBe(COMPACT);
  });

  it("hydrates a stored density on mount", () => {
    localStorage.setItem("data-density", COMFORTABLE);

    renderWithProviders(<DensityToggle />);

    expect(useDensityStore.getState().density).toBe(COMFORTABLE);
  });
});

describe("useDataDensity", () => {
  it("exposes the current density and reflects store changes", () => {
    const { result } = renderHookWithProviders(() => useDataDensity());

    expect(result.current).toBe(NORMAL);

    act(() => useDensityStore.getState().setDensity(COMFORTABLE));

    expect(result.current).toBe(COMFORTABLE);
  });
});
