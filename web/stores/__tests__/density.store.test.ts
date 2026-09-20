import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { COMFORTABLE, COMPACT, NORMAL, useDensityStore } from "../density.store";

describe("density store", () => {
  beforeEach(() => {
    localStorage.clear();
    useDensityStore.setState({ density: NORMAL });
  });

  afterEach(() => {
    useDensityStore.setState({ density: NORMAL });
  });

  it("starts at normal by default", () => {
    expect(useDensityStore.getState().density).toBe(NORMAL);
  });

  it("setDensity updates and persists the value", () => {
    useDensityStore.getState().setDensity(COMPACT);

    expect(useDensityStore.getState().density).toBe(COMPACT);
    expect(localStorage.getItem("data-density")).toBe(COMPACT);
  });

  it("hydrateDensity restores a stored value", () => {
    localStorage.setItem("data-density", COMFORTABLE);

    useDensityStore.getState().hydrateDensity();

    expect(useDensityStore.getState().density).toBe(COMFORTABLE);
  });

  it("hydrateDensity falls back to normal for invalid stored values", () => {
    localStorage.setItem("data-density", "extra-spacious");

    useDensityStore.getState().hydrateDensity();

    expect(useDensityStore.getState().density).toBe(NORMAL);
  });

  it("setDensity tolerates an unavailable localStorage", () => {
    const setItem = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("denied");
    });

    expect(() => useDensityStore.getState().setDensity(COMPACT)).not.toThrow();
    expect(useDensityStore.getState().density).toBe(COMPACT);
    setItem.mockRestore();
  });
});
